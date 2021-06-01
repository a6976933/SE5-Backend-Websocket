package wsHandler

import (
	"errors"
	"log"
)

const (
	ROOM_OPEN = iota
	ROOM_STOP
	ROOM_CLOSE
)

type RoomManager struct {
	roomCnt int
	//MessageID int
	LiveRoomList map[int]*RoomMsgManager
}

func NewRoomManager() *RoomManager {
	instance := new(RoomManager)
	instance.LiveRoomList = make(map[int]*RoomMsgManager)
	instance.roomCnt = 0 // using for test
	return instance
}

func (rm *RoomManager) CreateRoom(roomID int) (*RoomMsgManager, error) {
	newRoom := NewRoomMsgManager()
	newRoom.Manager = rm
	newRoom.ID = roomID
	rm.roomCnt += 1
	if _, ok := rm.LiveRoomList[roomID]; !ok {
		rm.LiveRoomList[roomID] = newRoom
		return newRoom, nil
	}
	log.Println("Room Exist Error")
	return nil, errors.New("Room Exist Error")
}

func (rm *RoomManager) IsRoomExist(roomID int) bool {
	if _, ok := rm.LiveRoomList[roomID]; ok {
		return true
	} else {
		return false
	}
}

func (rm *RoomManager) CloseRoom(roomID int) {
	delete(rm.LiveRoomList, roomID)
	log.Println("Close room, ID: ", roomID)
}
