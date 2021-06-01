package wsHandler

import (
	"errors"
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
)

const (
	HISTORY_MSG_NUM = 60
	LOAD_NUM        = 50
	MSG_SAVE_SIZE   = 70
	SAVE_TIME       = 120
	NOONE_TIME      = 180
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
	recTime     time.Time
	messageType string
	username    string
	userID      int
	roomID      int
	word        string
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

func (mq *MsgQueue) InitMsgQueue(msg []Msg, db *gorm.DB) {
	//db.Model(&RoomRoom{})
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
	ID               int    // room ID
	Name             string //room name
	Manager          *RoomManager
	OnlineMemberList map[int]*WsHandler
	UsernameMap      map[int]string
	AccessLevelMap   map[int]string
	historyMsgQueue  *MsgQueue
	broadcast        chan Msg
	register         chan WsRegister
	unregister       chan int
	message          chan Msg
	RoomInfo         *RoomRoom
	MessageSaveQueue []Room_Roommessage
	SaveMsgTicker    *time.Ticker
	NobodyTicker     *time.Ticker
}

func NewRoomMsgManager() *RoomMsgManager {
	instance := new(RoomMsgManager)
	instance.OnlineMemberList = make(map[int]*WsHandler)
	instance.UsernameMap = make(map[int]string)
	instance.AccessLevelMap = make(map[int]string)
	instance.MessageSaveQueue = make([]Room_Roommessage, 0)
	instance.broadcast = make(chan Msg)
	instance.register = make(chan WsRegister)
	instance.unregister = make(chan int)
	instance.message = make(chan Msg)
	instance.historyMsgQueue = NewMsgQueue()
	instance.SaveMsgTicker = time.NewTicker(SAVE_TIME * time.Second)
	instance.NobodyTicker = time.NewTicker(NOONE_TIME * time.Second)
	return instance
}

func (rmm *RoomMsgManager) LoadInitInfo(db *gorm.DB) error {
	count := int64(0)
	result := db.Model(&RoomRoom{}).Where("id = ?", strconv.Itoa(rmm.ID)).Count(&count)
	if count == 0 {
		log.Println("Can't find the room by room ID")
		return errors.New("Can't find the room by room ID")
	}
	if result.Error != nil {
		log.Println(result.Error)
		return result.Error
	}
	result = db.Preload("RoomMemberList").Preload("RoomBlockList").Find(rmm.RoomInfo, rmm.ID)
	if result.Error != nil {
		log.Println(result.Error)
		return result.Error
	}
	for _, item := range rmm.RoomInfo.RoomMemberList {
		rmm.UsernameMap[item.MemberID] = item.Nickname
		rmm.AccessLevelMap[item.MemberID] = item.AccessLevel
	}
	rmm.RoomInfo.RoomMemberList = nil
	rmm.Name = rmm.RoomInfo.Title
	result.Error = db.Model(&rmm.RoomInfo).Order("id desc").Limit(LOAD_NUM).Association("RoomMessageList").Find(&rmm.RoomInfo.RoomMessageList)

	if result.Error != nil {
		log.Println(result.Error)
		return result.Error
	}
	for i := 0; i < len(rmm.RoomInfo.RoomMessageList); i++ {
		loadMsg := TextMsg{
			recTime:     rmm.RoomInfo.RoomMessageList[i].RecvTime,
			messageType: "text",
			username:    rmm.UsernameMap[rmm.RoomInfo.RoomMessageList[i].MemberID],
			userID:      rmm.RoomInfo.RoomMessageList[i].MemberID,
			roomID:      rmm.ID,
			word:        rmm.RoomInfo.RoomMessageList[i].Message,
		}
		rmm.historyMsgQueue.Push(loadMsg)
	}
	return nil
}

func (rmm *RoomMsgManager) IsMemberInRoom(userID int) bool {
	if _, ok := rmm.AccessLevelMap[userID]; ok {
		return true
	} else {
		return false
	}
}

func (rmm *RoomMsgManager) SaveMsg2DBByNumber(db *gorm.DB) {
	if len(rmm.MessageSaveQueue) >= 70 {
		rmm.SaveMsg2DB(db)
	}
}

func (rmm *RoomMsgManager) SaveMsg2DBByTicker(db *gorm.DB) {
	rmm.SaveMsg2DB(db)
	rmm.SaveMsgTicker = time.NewTicker(SAVE_TIME * time.Second)
}

func (rmm *RoomMsgManager) SaveMsg2DB(db *gorm.DB) error {
	result := db.Create(&rmm.MessageSaveQueue)
	if result.Error != nil {
		log.Println(result.Error)
		return result.Error
	} else {
		rmm.MessageSaveQueue = rmm.MessageSaveQueue[:0]
	}
	return nil
}

func (rmm *RoomMsgManager) Run(db *gorm.DB) {
	for {
		select {
		case message := <-rmm.broadcast:
			log.Println("Got Broadcast ", message.GetTextMsg())
			log.Println("Now Online List", len(rmm.OnlineMemberList))
			if !rmm.historyMsgQueue.Push(message) {
				rmm.historyMsgQueue.RefreshHead()
				rmm.historyMsgQueue.Push(message)
			}
			if message.GetMsgType() == "text" {
				var tmpMsg = Room_Roommessage{RoomID: message.GetRoomID(), MemberID: message.GetUserID(), Message: message.GetTextMsg(), RecvTime: message.GetTime()}
				rmm.MessageSaveQueue = append(rmm.MessageSaveQueue, tmpMsg)
				log.Println(rmm.MessageSaveQueue)
			}
			rmm.SaveMsg2DBByNumber(db)

			for key, member := range rmm.OnlineMemberList {
				log.Println("UserID: ", key)
				log.Println("Member :", member.Username)
				select {
				case member.broadTextMsg <- message:
					log.Println("Send ", message.GetTextMsg(), " To ", member.Username)
				default:
					close(member.broadTextMsg)
					delete(rmm.OnlineMemberList, key)
				}
			}
		case member := <-rmm.register:
			log.Println(member.user.Username, " Register")
			rmm.OnlineMemberList[member.userID] = member.user
		case memberID := <-rmm.unregister:
			if _, ok := rmm.OnlineMemberList[memberID]; ok {
				log.Println(memberID, " Unregister")
				rmm.NobodyTicker = time.NewTicker(NOONE_TIME * time.Second)
				close(rmm.OnlineMemberList[memberID].broadTextMsg)
				delete(rmm.OnlineMemberList, memberID)
			}
		case <-rmm.SaveMsgTicker.C:
			rmm.SaveMsg2DBByTicker(db)
		case <-rmm.NobodyTicker.C:
			if len(rmm.OnlineMemberList) == 0 {
				rmm.SaveMsg2DB(db)
				rmm.Manager.CloseRoom(rmm.ID)
				return
			} else {
				rmm.NobodyTicker = time.NewTicker(NOONE_TIME * time.Second)
			}
		}
	}
}
