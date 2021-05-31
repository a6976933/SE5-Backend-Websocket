package wsHandler

import (
  //"gorm.io/gorm"
  "time"
)

type Room_Roommessage struct {
  ID int `gorm:"primaryKey"`
  RoomID int
  MemberID int
  Message string
  RecvTime time.Time
}

type UserCustomer struct {
  ID int `gorm:"primaryKey"`
  Password string
  LastLogin *time.Time `gorm:"type:datetime"`
  IsSuperuser bool
  Email string
  IsStaff bool
  IsActive bool
  DateJoined *time.Time `gorm:"type:datetime"`
  Username string
  FirstName string
  LastName string
  Department string
  Nickname string
  RoomMessageList []Room_Roommessage `gorm:"foreignKey:MemberID"`
  RoomMemberList []RoomRoommember `gorm:"foreignKey:MemberID"`
}

type RoomRoom struct {
  ID int `gorm:"primaryKey"`
  Title string
  Introduction string
  CreateTime *time.Time `gorm:"type:datetime"`
  ValidTime *time.Time `gorm:"type:datetime"`
  RoomType string
  RoomCategory string
  PeopleLimit int
  RoomBlockList []RoomRoomblock `gorm:"foreignKey:RoomID"`
  RoomInviteList []RoomRoominviting `gorm:"foreignKey:RoomID"`
  RoomInvitingList []RoomRoominvitingrequest `gorm:"foreignKey:RoomID"`
  RoomMemberList []RoomRoommember `gorm:"foreignKey:RoomID"`
  RoomRecordList []RoomRoomrecord `gorm:"foreignKey:RoomID"`
  RoomMessageList []Room_Roommessage `gorm:"foreignKey:RoomID"`
}

type RoomRoomblock struct {
  ID int `gorm:"primaryKey"`
  BlockTime *time.Time `gorm:"type:datetime"`
  Reason string
  BlockManagerID int
  BlockedUserID int
  RoomID int
}

type RoomRoominviting struct {
  ID int `gorm:"primaryKey"`
  InviteTime *time.Time `gorm:"type:datetime"`
  InvitedID int
  InviterID int
  RoomID int
}

type RoomRoominvitingrequest struct {
  ID int `gorm:"primaryKey"`
  RequestTime *time.Time `gorm:"type:datetime"`
  RequestUserID int
  RoomID int
}

type RoomRoommember struct {
  ID int `gorm:"primaryKey"`
  Nickname string
  AccessLevel string
  MemberID int
  RoomID int
}

type RoomRoomrecord struct {
  ID int `gorm:"primaryKey"`
  RecordTime *time.Time `gorm:"type:datetime"`
  Recording string
  RoomID int
}

type Tabler interface {
  TableName() string
}

func (Room_Roommessage) TableName() string {
  return "room_roommessage"
}

func (RoomRoom) TableName() string {
  return "room_room"
}

func (UserCustomer) TableName() string {
  return "user_customer"
}

func (RoomRoomblock) TableName() string {
  return "room_roomblock"
}

func (RoomRoominviting) TableName() string {
  return "room_roominviting"
}

func (RoomRoominvitingrequest) TableName() string {
  return "room_roominvitingrequest"
}

func (RoomRoommember) TableName() string {
  return "room_roommember"
}

func (RoomRoomrecord) TableName() string {
  return "room_roomrecord"
}
