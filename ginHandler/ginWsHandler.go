package ginHandler

import (
	"se5-back-websocket/wsHandler"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	//"golang.org/x/crypto/ssh"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"gorm.io/gorm"

	//"io/ioutil"
	"strconv"
	"time"
)

//openssl genrsa -out key.pem 2048
//openssl rsa -in key.pem -pubout -out key.pem.pub

const (
	//SECRETKEYPATH = "key.pem"
	MAX_AGE   = 300
	DJANGO_IP = "127.0.0.1"
)

type jwtRet struct {
	JWT string `json:"token"`
}

func RoomConnectHandler(rm *wsHandler.RoomManager, db *gorm.DB) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var roomID int
		var err error
		var servingRoom *wsHandler.RoomMsgManager
		var initInfo = new(wsHandler.RoomRequestConnectionMsg)
		wsH := wsHandler.NewWsHandler()
		if roomID, err = strconv.Atoi(c.Param("roomID")); err != nil {
			log.Println(err)
			c.JSON(http.StatusNotFound, gin.H{
				"detail": "Room Not Found",
			})
			return
		}
		if !rm.IsRoomExist(roomID) {
			count := int64(0)
			err := db.Model(&wsHandler.RoomRoom{}).Where("id = ?", roomID).Count(&count).Error
			if err != nil {
				log.Println(err)
				return
			}
			if count != 1 {
				log.Println("Access Room Error, Request room not Exist")
				return
			}
			servingRoom, err = rm.CreateRoom(roomID)
			if err != nil {
				return
			}
			servingRoom.ID = roomID
			servingRoom.RoomInfo = &wsHandler.RoomRoom{}
			err = servingRoom.LoadInitInfo(db)
			if err != nil {
				return
			}
		} else {
			servingRoom = rm.LiveRoomList[roomID]
			servingRoom.ID = roomID
		}
		go servingRoom.Run(db) //
		err = wsH.InitWebsocketConn(c)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"detail": "Websocket error",
			})
			return
		} else {
			log.Println("Someone connect to the room ID: ", roomID)
		}
		_, message, err := wsH.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		err = json.Unmarshal(message, &initInfo)
		if err != nil {
			log.Println(err)
			return
		} else {
			//log.Println("Has got the user information, \n name: ", initInfo.Username)
		}

		if _, ok := servingRoom.UsernameMap[initInfo.UserID]; !ok {
			log.Println("The user " + strconv.Itoa(initInfo.UserID) + " is not in the room, Invalid Access!!!!!")
			log.Println("So Close the Connection!!!")
			wsH.Conn.Close()
			return
		}

		isValid, JWTID := wsHandler.JWTAuthentication(initInfo.Token)
		if !isValid || JWTID != initInfo.UserID {
			wsH.Conn.Close()
			log.Println("Token Invalid!!!")
			return
		}

		wsH.UserID = initInfo.UserID
		wsH.Room = servingRoom
		wsH.Register()
		wsH.FetchMessage()
		go wsH.ReadPump()
		go wsH.WritePump()
	}
	return gin.HandlerFunc(fn)
}

func UserNotifyConnectionHandler(oum *wsHandler.OnlineUserManager, db *gorm.DB) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var userID int
		var err error
		var initInfo = new(wsHandler.OnlineUserRequestConnectionMsg)
		var UserOnline *wsHandler.OnlineUser
		if userID, err = strconv.Atoi(c.Param("id")); err != nil {
			log.Println(err)
			c.JSON(http.StatusNotFound, gin.H{
				"detail": "User Not Found",
			})
			return
		}
		if !oum.IsUserExist(userID) {
			count := int64(0)
			err := db.Model(&wsHandler.UserCustomuser{}).Where("id = ?", userID).Count(&count).Error
			if err != nil {
				log.Println(err)
				c.JSON(http.StatusNotFound, gin.H{
					"detail": "User Not Found",
				})
				return
			}
			if count != 1 {
				log.Println("Query User Error, Has two User")
				c.JSON(http.StatusInternalServerError, gin.H{
					"detail": "Query User Error",
				})
				return
			}
			UserOnline = wsHandler.NewOnlineUser()
			UserOnline.ID = userID
			UserOnline.UserInfo = &wsHandler.UserCustomuser{}
			err = UserOnline.LoadInitInfo(db)
			if err != nil {
				return
			}
			UserOnline.UserManager = oum
			UserOnline.Online = false
			UserOnline.Register()
		} else {
			UserOnline = oum.OnlineUserList[userID]
			UserOnline.ID = userID
			UserOnline.Online = true
		}
		err = UserOnline.InitWebsocketConn(c)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"detail": "Websocket error",
			})
			return
		} else {
			log.Println("User ID connect: ", UserOnline.ID)
		}
		_, message, err := UserOnline.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		err = json.Unmarshal(message, &initInfo)
		if err != nil {
			log.Println(err)
			return
		} else {
			//log.Println("Has got the user information, \n name: ", initInfo.UserID)
		}
		isValid, JWTID := wsHandler.JWTAuthentication(initInfo.Token)
		if !isValid || JWTID != UserOnline.ID {
			UserOnline.Conn.Close()
			return
		}
		err = oum.AddUser(UserOnline.ID, UserOnline)
		retMsg := []byte(`{"header": "setConn","res": "Success Connect"}`)
		UserOnline.Conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		err = UserOnline.Conn.WriteMessage(websocket.TextMessage, retMsg)
		if err != nil {
			log.Println(err)
			UserOnline.Conn.Close()
			return
		}
		log.Println("User ", UserOnline.ID, " Online!")
		if !UserOnline.Online {
			go UserOnline.TestConnection()
			UserOnline.Online = true
		}
		oum.UserCntReport()
	}
	return gin.HandlerFunc(fn)
}

func checkIP(c *gin.Context) error {
	if c.ClientIP() != DJANGO_IP {
		log.Println("Warning!!! Non server IP request our machine!!!")
		c.JSON(http.StatusNotFound, gin.H{
			"detail": "Page Not Found",
		})
		return errors.New("Warning!!! Non server IP request our machine!!!")
	}
	return nil
}

func RoomMemberRemoveHandler(oum *wsHandler.OnlineUserManager, rm *wsHandler.RoomManager, db *gorm.DB) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var roomID int
		err := checkIP(c)
		if err != nil {
			return
		}
		rmMessage := &wsHandler.RemoveMsg{}
		err = c.BindJSON(&rmMessage)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"detail": "Your JSON data format is wrong",
			})
		}
		roomID, err = strconv.Atoi(c.Param("id"))
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"detail": "Your ID is wrong",
			})
			return
		}
		modifiedRoom := rm.LiveRoomList[roomID]
		if modifiedRoom.IsMemberInRoom(rmMessage.RemovedUserID) {
			delete(modifiedRoom.AccessLevelMap, rmMessage.RemovedUserID)
		}
		c.JSON(http.StatusOK, gin.H{
			"detail": "Successful remove user from room",
		})
		//Sending Notification to front-end by WebSocket

	}
	return gin.HandlerFunc(fn)
}

func RoomMemberJoinHandler(oum *wsHandler.OnlineUserManager, rm *wsHandler.RoomManager, db *gorm.DB) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var roomID int
		err := checkIP(c)
		if err != nil {
			return
		}
		joinMessage := &wsHandler.JoinMsg{}
		err = c.BindJSON(&joinMessage)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"detail": "Your JSON data format is wrong",
			})
		}
		roomID, err = strconv.Atoi(c.Param("id"))
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"detail": "Your ID is wrong",
			})
			return
		}
		modifiedRoom := rm.LiveRoomList[roomID]
		if !modifiedRoom.IsMemberInRoom(joinMessage.JoinUserID) {
			modifiedRoom.AccessLevelMap[joinMessage.JoinUserID] = "user"
		}
		c.JSON(http.StatusOK, gin.H{
			"detail": "Successful join user from room",
		})
		//Sending Notification to front-end by WebSocket

	}
	return gin.HandlerFunc(fn)
}

func RoomUpdateHandler(oum *wsHandler.OnlineUserManager, rm *wsHandler.RoomManager, db *gorm.DB) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var roomID int
		err := checkIP(c)
		if err != nil {
			return
		}
		updateMessage := &wsHandler.UpdateMsg{}
		err = c.BindJSON(&updateMessage)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"detail": "Your JSON data format is wrong",
			})
		}
		roomID, err = strconv.Atoi(c.Param("id"))
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"detail": "Your ID is wrong",
			})
			return
		}
		modifiedRoom := rm.LiveRoomList[roomID]
		c.JSON(http.StatusOK, gin.H{
			"detail": "Successful update",
		})
		log.Println(modifiedRoom.ID)
	}
	return gin.HandlerFunc(fn)
}

func BackendUserNotifyHandler(oum *wsHandler.OnlineUserManager, db *gorm.DB) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var userID int
		//var user *wsHandler.OnlineUser
		err := checkIP(c)
		if err != nil {
			return
		}
		userNotify := &wsHandler.UserNotificationMsg{}
		err = c.BindJSON(&userNotify)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"detail": "Your JSON data format is wrong",
			})
		}
		userID, err = strconv.Atoi(c.Param("id"))
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"detail": "Your ID is wrong",
			})
			return
		}
		if _, ok := oum.OnlineUserList[userID]; !ok {
			c.JSON(http.StatusOK, gin.H{
				"detail": "User isn't online",
			})
			return
		}
		userNotify.NotifyUserID = userID
		c.JSON(http.StatusOK, gin.H{
			"detail": "Send Success",
		})
		oum.Notify <- *userNotify
	}
	return gin.HandlerFunc(fn)
}
