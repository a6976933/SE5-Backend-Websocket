package main

import (
	"log"
	"se5-back-websocket/wsHandler"
	"strconv"
	"testing"
)

func Test_Room(t *testing.T) {
	//dsn := "host=ec2-3-218-71-191.compute-1.amazonaws.com user=ktpcfcntkcpxwi password=71db68e37c141279bb86464e2df66f83e184b5f28c8911f2da498c6fb0aa482d dbname=d1ictoavo8addi port=5432 sslmode=require"
	//db, err := initDatabase(dsn)
	roomManager := wsHandler.NewRoomManager()
	roomID := 3
	_, err := roomManager.CreateRoom(roomID)
	_, err = roomManager.CreateRoom(roomID)
	if err == nil {
		t.Error("Room duplicate")
	} else {
		log.Println(err)
	}
}

func Test_UserJoin_In_Room(t *testing.T) {
	dsn := "host=ec2-3-218-71-191.compute-1.amazonaws.com user=ktpcfcntkcpxwi password=71db68e37c141279bb86464e2df66f83e184b5f28c8911f2da498c6fb0aa482d dbname=d1ictoavo8addi port=5432 sslmode=require"
	db, err := initDatabase(dsn)
	roomManager := wsHandler.NewRoomManager()
	roomID := 7
	servingRoom, err := roomManager.CreateRoom(roomID)
	if err != nil {
		t.Error(err)
		return
	}
	servingRoom.ID = roomID
	servingRoom.RoomInfo = &wsHandler.RoomRoom{}
	roomExist := true
	err = servingRoom.LoadInitInfo(db)
	if err != nil {
		t.Error(err)
		return
	}
	memberID := 4
	if roomExist {
		if _, ok := servingRoom.UsernameMap[memberID]; !ok {
			var notInListUser wsHandler.RoomRoommember
			result := db.Where("room_id = ? AND member_id = ?", roomID, memberID).First(&notInListUser)
			if result.Error != nil {
				log.Println("The user " + strconv.Itoa(memberID) + " is not in the room, Invalid Access!!!!!")
				log.Println("So Close the Connection!!!")
				t.Error("Error! Should find user")
				return
			}
		}
	}
	for k, v := range servingRoom.UsernameMap {
		log.Println(k, v)
	}
}

func Test_UserJoin_Not_In_Room(t *testing.T) {
	dsn := "host=ec2-3-218-71-191.compute-1.amazonaws.com user=ktpcfcntkcpxwi password=71db68e37c141279bb86464e2df66f83e184b5f28c8911f2da498c6fb0aa482d dbname=d1ictoavo8addi port=5432 sslmode=require"
	db, err := initDatabase(dsn)
	roomManager := wsHandler.NewRoomManager()
	roomID := 1
	servingRoom, err := roomManager.CreateRoom(roomID)
	if err != nil {
		t.Error(err)
		return
	}
	servingRoom.ID = roomID
	servingRoom.RoomInfo = &wsHandler.RoomRoom{}
	roomExist := true
	err = servingRoom.LoadInitInfo(db)
	if err != nil {
		t.Error(err)
		return
	}
	memberID := 8
	if roomExist {
		if _, ok := servingRoom.UsernameMap[memberID]; !ok {
			var notInListUser wsHandler.RoomRoommember
			result := db.Where("room_id = ? AND member_id = ?", roomID, memberID).First(&notInListUser)
			if result.Error != nil {
				log.Println("The user " + strconv.Itoa(memberID) + " is not in the room, Invalid Access!!!!!")
				log.Println("So Close the Connection!!!")
				return
			} else {
				t.Error("Error! User Should be block!!!")
			}
		}
	}
	for k, v := range servingRoom.UsernameMap {
		log.Println(k, v)
	}
}

func Test_UserJoin_Join_Room(t *testing.T) {
	dsn := "host=ec2-3-218-71-191.compute-1.amazonaws.com user=ktpcfcntkcpxwi password=71db68e37c141279bb86464e2df66f83e184b5f28c8911f2da498c6fb0aa482d dbname=d1ictoavo8addi port=5432 sslmode=require"
	db, err := initDatabase(dsn)
	roomManager := wsHandler.NewRoomManager()
	roomID := 1
	servingRoom, err := roomManager.CreateRoom(roomID)
	if err != nil {
		t.Error(err)
		return
	}
	servingRoom.ID = roomID
	servingRoom.RoomInfo = &wsHandler.RoomRoom{}
	roomExist := true
	err = servingRoom.LoadInitInfo(db)
	if err != nil {
		t.Error(err)
		return
	}
	memberID := 5
	delete(servingRoom.UsernameMap, 5)
	if roomExist {
		if _, ok := servingRoom.UsernameMap[memberID]; !ok {
			var notInListUser wsHandler.RoomRoommember
			result := db.Where("room_id = ? AND member_id = ?", roomID, memberID).First(&notInListUser)
			if result.Error != nil {
				t.Error("Error! User should join the room !!!")
				log.Println("The user " + strconv.Itoa(memberID) + " is not in the room, Invalid Access!!!!!")
				log.Println("So Close the Connection!!!")
				return
			} else {

			}
		}
	}
	for k, v := range servingRoom.UsernameMap {
		log.Println(k, v)
	}
}

func Test_update_not_exist_room(t *testing.T) {
	dsn := "host=ec2-3-218-71-191.compute-1.amazonaws.com user=ktpcfcntkcpxwi password=71db68e37c141279bb86464e2df66f83e184b5f28c8911f2da498c6fb0aa482d dbname=d1ictoavo8addi port=5432 sslmode=require"
	db, err := initDatabase(dsn)
	roomManager := wsHandler.NewRoomManager()
	roomID := 1
	servingRoom, err := roomManager.CreateRoom(roomID)
	if err != nil {
		t.Error(err)
		return
	}
	servingRoom.ID = roomID
	servingRoom.RoomInfo = &wsHandler.RoomRoom{}
	//roomExist := true
	err = servingRoom.LoadInitInfo(db)
	if err != nil {
		t.Error(err)
		return
	}
	if !roomManager.IsRoomExist(1123) {
		return
	} else {
		t.Error("Update doesn't exist room!")
	}
}

func Test_delete_room_forgery(t *testing.T) {
	dsn := "host=ec2-3-218-71-191.compute-1.amazonaws.com user=ktpcfcntkcpxwi password=71db68e37c141279bb86464e2df66f83e184b5f28c8911f2da498c6fb0aa482d dbname=d1ictoavo8addi port=5432 sslmode=require"
	db, err := initDatabase(dsn)
	roomManager := wsHandler.NewRoomManager()
	roomID := 1
	servingRoom, err := roomManager.CreateRoom(roomID)
	if err != nil {
		t.Error(err)
		return
	}
	servingRoom.ID = roomID
	servingRoom.RoomInfo = &wsHandler.RoomRoom{}
	err = servingRoom.LoadInitInfo(db)
	if err != nil {
		t.Error(err)
		return
	}
	//modifiedRoom := servingRoom
	cnt := int64(0)
	db.Model(&wsHandler.RoomRoom{}).Where("id = ?", roomID).Count(&cnt)
	if cnt > 0 {
		return
	} else {
		t.Error("Room delete has been forgeried!!!")
	}
}
