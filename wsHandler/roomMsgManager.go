package wsHandler

import (
  "time"
  "log"
)

const (
  HISTORY_MSG_NUM = 50
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

type MsgQueue struct {
  queue []Msg
}

func NewMsgQueue() *MsgQueue {
  instance := new(MsgQueue)
  instance.queue = make([]Msg, 0)
  return instance
}

func (mq *MsgQueue) InitMsgQueue(msg []Msg) {
  // Wait ORM Fetching
}

func (mq *MsgQueue) IsEmpty() bool {
  if len(mq.queue) == 0 {
    return true
  } else {
    return false
  }
}

func (mq *MsgQueue) Push(msg Msg) bool {
  if len(mq.queue) < HISTORY_MSG_NUM {
    mq.queue = append(mq.queue, msg)
    return true
  } else {
    return false
  }
}

func (mq *MsgQueue) PopHead() Msg {
  msg := mq.queue[0]
  mq.queue = mq.queue[1:]
  return msg
}

func (mq *MsgQueue) FetchAll() []Msg {
  ret := make([]Msg, 0)
  for _, v := range mq.queue {
    ret = append(ret, v)
  }
  return ret
}

func (mq *MsgQueue) RefreshHead() {
  mq.queue = mq.queue[1:]
}


type RoomMsgManager struct {
  ID int // room ID
  Name string //room name
  Manager *RoomManager
  OnlineMemberList map[int]*WsHandler
  historyMsgQueue *MsgQueue
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
  instance.historyMsgQueue = NewMsgQueue()
  return instance
}

func (rmm *RoomMsgManager) Run() {
  for {
    select {
    case message := <-rmm.broadcast:
      log.Println("Got Broadcast ", message.GetTextMsg())
      log.Println("Now Online List", len(rmm.OnlineMemberList))
      if !rmm.historyMsgQueue.Push(message) {
        rmm.historyMsgQueue.RefreshHead()
        rmm.historyMsgQueue.Push(message)
      }

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
