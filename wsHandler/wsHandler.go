package wsHandler

import (
	"encoding/json"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	//"io/ioutil"
	//"golang.org/x/crypto/ssh"
	//"strings"
	//"strconv"
	//"math/big"
	//"crypto/rsa"
	"errors"
	"log"
	"net/http"
	"time"
)

const (
	KEY           = "django-insecure-wn1h_@!bp!zbv5lm9dwh63m$hf#bvy+u#ef+i&y3m!&7nw(^15"
	PUBLICKEYPATH = "key.pem.pub"
)

type JWTClaim struct {
	UserID int
	//Username string
	jwt.StandardClaims
}

type WsRegister struct {
	userID int
	user   *WsHandler
}

type WsHandler struct {
	Upgrader     *websocket.Upgrader
	Conn         *websocket.Conn
	Room         *RoomMsgManager
	broadTextMsg chan Msg
	Username     string
	UserID       int
}

type RecvMsg struct {
	UserID     int    `json:"userID"`
	Username   string `json:"user_name"`
	MsgType    string `json:"msg_type"`
	RoomName   string `json:"room_name"`
	RoomID     int    `json:"roomID"`
	ImgMessage []byte `json:"img_message"`
	Message    string `json:"message"`
	JWTToken   string `json:"jwt"`
}

type WriteMsg struct {
	UserID     int    `json:"userID"`
	Username   string `json:"username"`
	MsgType    string `json:"msg_type"`
	RoomName   string `json:"room_name"`
	RoomID     int    `json:"roomID"`
	ImgMessage []byte `json:"img_message"`
	Message    string `json:"message"`
}

const (
	maxMsgReadSize = 2048
	writeWait      = 1 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
)

func (wsh *WsHandler) InitWebsocketConn(c *gin.Context) error {
	var err error
	wsh.Conn, err = wsh.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		wsh.Conn = nil
		return err
	}
	return nil
}

func (wsh *WsHandler) initUpgrader(rbSize int, wbSize int) {
	wsh.Upgrader = &websocket.Upgrader{
		ReadBufferSize:  rbSize,
		WriteBufferSize: wbSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
}

func (wsh *WsHandler) FetchMessage() bool {
	if !wsh.Room.historyMsgQueue.IsEmpty() {
		historyMsg := wsh.Room.historyMsgQueue.FetchAll()
		for i := 0; i < len(historyMsg); i++ {
			wsh.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			writeMessage := wsh.SetWriteMsg(historyMsg[i])
			sendMsg, _ := json.Marshal(writeMessage)
			err := wsh.Conn.WriteMessage(websocket.TextMessage, sendMsg)
			if err != nil {
				wsh.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				log.Println("Write history message to member ", wsh.Username, " False")
				return false
			}
		}
	} else {
		return false
	}
	return true
}

func (wsh *WsHandler) Register() {
	regInfo := WsRegister{userID: wsh.UserID, user: wsh}
	wsh.Room.register <- regInfo
}

func (wsh *WsHandler) ReadPump() {
	defer func() {
		wsh.Conn.Close()
		wsh.Room.unregister <- wsh.UserID
	}()
	wsh.Conn.SetReadDeadline(time.Now().Add(pongWait))
	wsh.Conn.SetPongHandler(func(string) error { wsh.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		var textMessage TextMsg
		var recvMessage RecvMsg
		_, message, err := wsh.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			} else {
				log.Println(err)
			}
			break
		}
		err = json.Unmarshal(message, &recvMessage)
		if err != nil {
			log.Println(err)
			break
		}
		log.Println("JWT Token: ", recvMessage.JWTToken)
		parsetoken, err := jwt.ParseWithClaims(recvMessage.JWTToken, &JWTClaim{}, func(token *jwt.Token) (interface{}, error) {
			//var err error
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("Token Algorithm wrong")
			}
			if token.Header["typ"] != "JWT" || token.Header["alg"] != "HS256" {
				return nil, errors.New("Expected typ JWT and alg HS256")
			}
			/*
			   pubkeyFile , err := ioutil.ReadFile(PUBLICKEYPATH)
			   if err != nil {
			     log.Println(err," Read Error")
			     return nil, errors.New(err.Error()+" Read Error")
			   }
			   pubkey, err := jwt.ParseRSAPublicKeyFromPEM(pubkeyFile)
			   if err != nil {
			     log.Println(err," Parse Error")
			     return nil, errors.New(err.Error()+" Parse Error")
			   }*/ // RS256 code
			key := []byte(KEY)
			return key, nil
		})
		if err != nil {
			log.Println(err, " Parse Token Error")
			return
		}
		if !parsetoken.Valid {
			log.Println("Token is invalid")
			return
		} else {
			log.Println("Token is valid")
		}
		jwtInfo := parsetoken.Claims.(*JWTClaim)
		//log.Println("Name: ", jwtInfo.Username)
		log.Println("ID: ", jwtInfo.UserID)
		log.Println("Username: ", recvMessage.Username, "Message: ", recvMessage.Message, "Room Name: ", recvMessage.RoomName)
		if recvMessage.MsgType == "text" {
			textMessage.userID = recvMessage.UserID
			textMessage.roomID = recvMessage.RoomID
			textMessage.username = recvMessage.Username
			textMessage.word = recvMessage.Message
			textMessage.messageType = "text"
			textMessage.recTime = time.Now()
			wsh.Room.broadcast <- textMessage
		}

	}

}

func (wsh *WsHandler) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	defer wsh.Conn.Close()
	for {
		select {
		case message, ok := <-wsh.broadTextMsg:
			wsh.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				wsh.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if message.GetMsgType() == "text" {
				wMsg := wsh.SetWriteMsg(message)
				sendMsg, _ := json.Marshal(wMsg)
				log.Println("Send JSON to client: ", string(sendMsg))
				err := wsh.Conn.WriteMessage(websocket.TextMessage, sendMsg)
				if err != nil {
					wsh.Conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
			}
		case <-ticker.C:
			wsh.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := wsh.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (wsh *WsHandler) SetWriteMsg(message Msg) WriteMsg {
	var wMsg WriteMsg
	if message.GetMsgType() == "text" {
		wMsg.Message = message.GetTextMsg()
		wMsg.MsgType = message.GetMsgType()
		wMsg.UserID = message.GetUserID()
		wMsg.Username = message.GetUsername()
		wMsg.RoomID = message.GetRoomID()
	}
	return wMsg
}

func NewWsHandler() *WsHandler {
	instance := new(WsHandler)
	rbSize := 2048
	wbSize := 2048
	instance.initUpgrader(rbSize, wbSize)
	instance.broadTextMsg = make(chan Msg)
	return instance
}
