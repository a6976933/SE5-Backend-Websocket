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

	"log"
	"net/http"
	"time"
)

const (
	KEY           = "django-insecure-wn1h_@!bp!zbv5lm9dwh63m$hf#bvy+u#ef+i&y3m!&7nw(^15"
	PUBLICKEYPATH = "key.pem.pub"
)

type JWTClaim struct {
	UserID int `json:"user_id"`
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
	Nickname     string
	UserID       int
	RoomID       int
}

type RoomRequestConnectionMsg struct {
	UserID int    `json:"user_id"`
	RoomID int    `json:"room_id"`
	Token  string `json:"token"`
}

type RecvMsg struct {
	Header     string `json:"header"`
	UserID     int    `json:"userID"`
	Username   string `json:"user_name"`
	MsgType    string `json:"msg_type"`
	RoomName   string `json:"room_name"`
	RoomID     int    `json:"roomID"`
	ImgMessage []byte `json:"img_message"`
	Message    string `json:"message"`
	JWTToken   string `json:"token"`
}

type WriteMsg struct {
	Header     string    `json:"header"`
	UserID     int       `json:"userID"`
	Nickname   string    `json:"nickname"`
	MsgType    string    `json:"msg_type"`
	RoomName   string    `json:"room_name"`
	RoomID     int       `json:"roomID"`
	MsgTime    time.Time `json:"time"`
	ImgMessage []byte    `json:"img_message"`
	Message    string    `json:"message"`
}

const (
	maxMsgReadSize = 2048
	writeWait      = 1 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 20 //(pongWait * 9) / 10
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
				log.Println("Write history message to member ", wsh.UserID, " got error, so close the connection")
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

func (wsh *WsHandler) Unregister() {
	wsh.Room.unregister <- wsh.UserID
}

func (wsh *WsHandler) ReadPump() {
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
		log.Println(recvMessage)
		isValid, JWTID := JWTAuthentication(recvMessage.JWTToken)
		if !isValid || JWTID != recvMessage.UserID {
			wsh.Conn.Close()
			log.Println("Token Invalid!!!", JWTID, recvMessage.UserID)
			wsh.Unregister()
			return
		}
		if recvMessage.Header == "message" {
			if recvMessage.MsgType == "text" {
				textMessage.userID = recvMessage.UserID
				textMessage.roomID = wsh.RoomID
				textMessage.username = wsh.Nickname
				textMessage.word = recvMessage.Message
				textMessage.messageType = "text"
				textMessage.recTime = time.Now()
				wsh.Room.broadcast <- textMessage
			}
		} else if recvMessage.Header == "controlMessage" {
			if recvMessage.Message == "Close" {
				wsh.Unregister()
				return
			}
		}

	}

}

func (wsh *WsHandler) WritePump() {
	ticker := time.NewTicker(pingPeriod * time.Second)
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
			wMsg := wsh.SetWriteMsg(message)
			sendMsg, _ := json.Marshal(wMsg)
			log.Println("Send JSON to client: ", string(sendMsg))
			err := wsh.Conn.WriteMessage(websocket.TextMessage, sendMsg)
			if err != nil {
				wsh.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

		case <-ticker.C:
			wsh.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			msg := []byte(`{"header":"ping","ping": "ping"}`)
			if err := wsh.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				wsh.Room.unregister <- wsh.UserID
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
		wMsg.Nickname = message.GetUsername()
		wMsg.RoomID = message.GetRoomID()
		wMsg.MsgTime = message.GetTime()
		wMsg.Header = "message"
	} else if message.GetMsgType() == "update" {
		wMsg.Message = message.GetTextMsg()
		wMsg.MsgType = message.GetMsgType()
		wMsg.RoomID = message.GetRoomID()
		wMsg.MsgTime = message.GetTime()
		wMsg.Header = message.GetMsgHeader()
	} else if message.GetMsgType() == "remove" {
		wMsg.Message = message.GetTextMsg()
		wMsg.MsgType = message.GetMsgType()
		wMsg.RoomID = message.GetRoomID()
		wMsg.MsgTime = message.GetTime()
		wMsg.Header = message.GetMsgHeader()
		wMsg.UserID = message.GetUserID()
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
