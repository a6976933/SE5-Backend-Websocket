package wsHandler

import (
  "time"
  "log"
)

type Msg interface {
  GetTime() time.Time
  GetUsername() string
  GetUserID() int
  GetRoomID() int
  GetMsgType() string
  GetTextMsg() string
  //GetJPGMsg()
}

type TextMsg struct {
  recTime time.Time
  messageType string
  username string
  userID int
  roomID int
  word string
}

func (tm TextMsg) GetTime() time.Time {
  return tm.recTime
}

func (tm TextMsg) GetUsername() string {
  return tm.username
}

func (tm TextMsg) GetUserID() int {
  return tm.userID
}

func (tm TextMsg) GetRoomID() int {
  return tm.roomID
}

func (tm TextMsg) GetMsgType() string {
  return tm.messageType
}

func (tm TextMsg) GetTextMsg() string {
  return tm.word
}

type RoomMsgManager struct {
  ID int // room ID
  Name string //room name
  Manager *RoomManager
  OnlineMemberList map[int]*WsHandler
  broadcast chan Msg
  register chan WsRegister
  unregister chan int
  message chan Msg
}

func NewRoomMsgManager() *RoomMsgManager {
  instance := new(RoomMsgManager)
  instance.OnlineMemberList = make(map[int]*WsHandler)
  instance.broadcast = make(chan Msg)
  instance.register = make(chan WsRegister)
  instance.unregister = make(chan int)
  instance.message = make(chan Msg)
  return instance
}

func (rmm *RoomMsgManager) Run() {
  for {
    select {
    case message := <-rmm.broadcast:
      log.Println("Got Broadcast ", message.GetTextMsg())
      log.Println("Now Online List", len(rmm.OnlineMemberList))
      for key, member := range rmm.OnlineMemberList {
        log.Println("UserID: ",key)
        log.Println("Member :", member.Username)
        select {
        case member.broadTextMsg <- message:
          log.Println("Send ",message.GetTextMsg()," To ", member.Username)
        default:
          close(member.broadTextMsg)
          delete(rmm.OnlineMemberList, key)
        }
      }
    case member := <-rmm.register:
      log.Println(member.user.Username," Register")
      rmm.OnlineMemberList[member.userID] = member.user
    case memberID := <-rmm.unregister:
      if _, ok := rmm.OnlineMemberList[memberID]; ok {
        log.Println(memberID, " Unregister")
        close(rmm.OnlineMemberList[memberID].broadTextMsg)
        delete(rmm.OnlineMemberList, memberID)
      }

    }
  }
}
