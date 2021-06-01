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

	"github.com/dgrijalva/jwt-go"
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
		var initInfo = new(wsHandler.RecvMsg)
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
			log.Println("Has got the user information, \n name: ", initInfo.Username)
		}

		if _, ok := servingRoom.UsernameMap[initInfo.UserID]; !ok {
			log.Println("The user " + string(initInfo.UserID) + " is not in the room")
		}

		key := []byte(wsHandler.KEY)
		if err != nil {
			log.Println(err)
		}

		//-----------
		customClaim := &wsHandler.JWTClaim{
			UserID: initInfo.UserID,
			//Username: initInfo.Username,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Duration(MAX_AGE) * time.Second).Unix(),
				Issuer:    initInfo.Username,
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaim)
		tokenString, err := token.SignedString(key)
		log.Println(tokenString)
		jwtObj := &jwtRet{JWT: tokenString}
		sendJWT, _ := json.Marshal(jwtObj)
		wsH.Conn.WriteMessage(websocket.TextMessage, sendJWT)
		if err != nil {
			log.Println(err)
		}
		wsH.Username = initInfo.Username
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
			err := db.Model(&wsHandler.UserCustomer{}).Where("id = ?", userID).Count(&count).Error
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
			UserOnline.UserInfo = &wsHandler.UserCustomer{}
			err = UserOnline.LoadInitInfo(db)
			if err != nil {
				return
			}
		} else { // Usually can't go to here
			UserOnline = oum.OnlineUserList[userID]
			UserOnline.ID = userID
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
		if err != nil {
			return
		}
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
