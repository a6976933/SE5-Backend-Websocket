package wsHandler

import (
  "github.com/gorilla/websocket"
  "github.com/gin-gonic/gin"
  "encoding/json"
  "net/http"
  "log"
  "time"
)

type WsRegister struct {
  userID int
  user *WsHandler
}

type WsHandler struct {
  Upgrader *websocket.Upgrader
  Conn *websocket.Conn
  Room *RoomMsgManager
  broadTextMsg chan Msg
  Username string
  UserID int
}

type RecvMsg struct {
  UserID int `json:"userID"`
  Username string `json:"user_name"`
  MsgType string `json:"msg_type"`
  RoomName string `json:"room_name"`
  RoomID int `json:"roomID"`
  ImgMessage []byte `json:"img_message"`
  Message string `json:"message"`
}

type WriteMsg struct {
  UserID int `json:"userID"`
  Username string `json:"username"`
  MsgType string `json:"msg_type"`
  RoomName string `json:"room_name"`
  RoomID int `json:"roomID"`
  ImgMessage []byte `json:"img_message"`
  Message string `json:"message"`
}

const (
  maxMsgReadSize = 2048

  writeWait = 1 * time.Second

  pongWait = 60 * time.Second

  pingPeriod = (pongWait * 9) / 10
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
    ReadBufferSize: rbSize,
    WriteBufferSize: wbSize,
    CheckOrigin: func(r *http.Request) bool {
  		return true
  	},
  }
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
  wsh.Conn.SetPongHandler(func(string) error { wsh.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil})

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
        var wMsg WriteMsg
        wMsg.Message = message.GetTextMsg()
        wMsg.MsgType = message.GetMsgType()
        wMsg.UserID = message.GetUserID()
        wMsg.Username = message.GetUsername()
        wMsg.RoomID = message.GetRoomID()

        sendMsg, _ := json.Marshal(wMsg)
        log.Println("Send JSON to client: ",string(sendMsg))
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

func NewWsHandler() *WsHandler {
  instance := new(WsHandler)
  rbSize := 2048
  wbSize := 2048
  instance.initUpgrader(rbSize, wbSize)
  instance.broadTextMsg = make(chan Msg)
  return instance
}
