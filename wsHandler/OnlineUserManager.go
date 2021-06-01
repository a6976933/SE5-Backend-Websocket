package wsHandler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type RemoveMsg struct {
	RemoveUserID  int `json:"remove_userID"`
	RemovedUserID int `json:"removed_userID"`
}

type JoinMsg struct {
	JoinUserID int `json:"join_userID"`
}

type UpdateMsg struct {
	UpdateItem string `json:"update_item"`
	UpdateData string `json:"update_data"`
}

type UserNotificationMsg struct {
	NotifyType     string `json:"notify_type"`
	NotifyString   string `json:"notify_string"`
	DoNotifyUserID int    `json:"notify_userID"`
	NotifyRoomID   int    `json:"notify_roomID"`
	NotifyUserID   int    `json:"notified_userID"`
}

type OnlineUserManager struct {
	onlineUserCnt  int
	register       chan OnlineUserRegister
	unregister     chan int
	Notify         chan UserNotificationMsg
	OnlineUserList map[int]*OnlineUser
}

func (oum *OnlineUserManager) AddUser(userID int, user *OnlineUser) error {
	if _, ok := oum.OnlineUserList[userID]; !ok {
		oum.OnlineUserList[userID] = user
		return nil
	}
	log.Println("User Exist Error")
	return errors.New("User Exist Error")
}

func (oum *OnlineUserManager) Run() {
	for {
		select {
		case message := <-oum.Notify:
			user := oum.OnlineUserList[message.NotifyUserID]
			go func(user *OnlineUser, msg UserNotificationMsg) {
				ticker := time.NewTicker(pingPeriod)
				defer ticker.Stop()
				marshMsg, _ := json.Marshal(msg)
				err := user.Conn.WriteMessage(websocket.TextMessage, marshMsg)
				if err != nil {
					user.Conn.WriteMessage(websocket.CloseMessage, []byte{})
					user.Unregister()
				}
			}(user, message)
		case userID := <-oum.unregister:
			if userObj, ok := oum.OnlineUserList[userID]; ok {
				userObj.Conn.Close()
				close(userObj.notifyMsg)
				delete(oum.OnlineUserList, userID)
			}
		case userReg := <-oum.register:
			oum.OnlineUserList[userReg.userID] = userReg.user
		}
	}
}

func (oum *OnlineUserManager) IsUserExist(userID int) bool {
	if _, ok := oum.OnlineUserList[userID]; ok {
		return true
	} else {
		return false
	}
}

func NewOnlineUserManager() *OnlineUserManager {
	instance := new(OnlineUserManager)
	instance.OnlineUserList = make(map[int]*OnlineUser)
	instance.onlineUserCnt = 0
	instance.Notify = make(chan UserNotificationMsg)
	instance.register = make(chan OnlineUserRegister)
	instance.unregister = make(chan int)
	return instance
}

type OnlineUserRequestConnectionMsg struct {
	UserID int    `json:"user_id"`
	Token  string `json:"token"`
}

type OnlineUserRegister struct {
	userID int
	user   *OnlineUser
}

type OnlineUser struct {
	Upgrader    *websocket.Upgrader
	Conn        *websocket.Conn
	UserManager *OnlineUserManager
	UserInfo    *UserCustomer
	notifyMsg   chan Msg
	Username    string
	ID          int
}

func NewOnlineUser() *OnlineUser {
	instance := new(OnlineUser)
	rbSize := 2048
	wbSize := 2048
	instance.initUpgrader(rbSize, wbSize)
	instance.notifyMsg = make(chan Msg)
	return instance
}

func (ou *OnlineUser) LoadInitInfo(db *gorm.DB) error {
	count := int64(0)
	result := db.Model(&UserCustomer{}).Where("id = ?", strconv.Itoa(ou.ID)).Count(&count)
	if count == 0 {
		log.Println("Can't find the user by user ID")
		return errors.New("Can't find the user by user ID")
	}
	if result.Error != nil {
		log.Println(result.Error)
		return result.Error
	}
	result = db.Preload("RoomMemberList").Find(&ou.UserInfo, ou.ID)
	if result.Error != nil {
		log.Println(result.Error)
		return result.Error
	}
	ou.Username = ou.UserInfo.Username
	return nil
}

func (ou *OnlineUser) initUpgrader(rbSize int, wbSize int) {
	ou.Upgrader = &websocket.Upgrader{
		ReadBufferSize:  rbSize,
		WriteBufferSize: wbSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
}

func (ou *OnlineUser) InitWebsocketConn(c *gin.Context) error {
	var err error
	ou.Conn, err = ou.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		ou.Conn = nil
		return err
	}
	return nil
}

func (ou *OnlineUser) Register() {
	regInfo := OnlineUserRegister{userID: ou.ID, user: ou}
	ou.UserManager.register <- regInfo
}

func (ou *OnlineUser) Unregister() {
	ou.UserManager.unregister <- ou.ID
}

func JWTAuthentication(token string) (bool, int) {
	parsetoken, err := jwt.ParseWithClaims(token, &JWTClaim{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Token Algorithm wrong")
		}
		if token.Header["typ"] != "JWT" || token.Header["alg"] != "HS256" {
			return nil, errors.New("Expected typ JWT and alg HS256")
		}
		key := []byte(KEY)
		return key, nil
	})
	if err != nil {
		log.Println(err, " Parse Token Error")
		return false, -1
	}
	if !parsetoken.Valid {
		log.Println("Token is invalid")
		return false, -1
	}
	jwtInfo := parsetoken.Claims.(*JWTClaim)
	log.Println("Token is valid, User ID: ", jwtInfo.UserID)
	return true, jwtInfo.UserID
}
