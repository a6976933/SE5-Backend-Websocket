package main

import (
	"log"
	"se5-back-websocket/ginHandler"
	"se5-back-websocket/wsHandler"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	//"strconv"
	//"time"
	"gorm.io/driver/postgres"
	//"gorm.io/driver/sqlite"
)

func initDatabase(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return db, nil
}

func main() {
	const addr = "127.0.0.1:8090"
	//gin.SetMode(gin.ReleaseMode)

	var dsn string //"/Users/allenwang/se5-back/db.sqlite3"
	dsn = "host=ec2-3-218-71-191.compute-1.amazonaws.com user=ktpcfcntkcpxwi password=71db68e37c141279bb86464e2df66f83e184b5f28c8911f2da498c6fb0aa482d dbname=d1ictoavo8addi port=5432 sslmode=require"
	//fmt.Println("input DSN(Sqlite location)")
	//fmt.Scanln(&dsn)
	db, err := initDatabase(dsn)
	if err != nil {
		return
	}
	var messageID = int64(0)
	err = db.Model(&wsHandler.Room_Roommessage{}).Count(&messageID).Error
	if err != nil {
		log.Println(err)
	}
	/*
		log.Println(messageID)
		for i := 1; i < 10; i++ {
			test := wsHandler.Room_Roommessage{RoomID: 4, MemberID: 2, Message: "上課囉AAAAAA" + strconv.Itoa(i), RecvTime: time.Now(), Nickname: "allen"}
			//messageID += 1
			result := db.Create(&test)
			if result.Error != nil {
				log.Println(result.Error)
				return
			}
		}*/
	roomManager := wsHandler.NewRoomManager()
	userManager := wsHandler.NewOnlineUserManager()
	go userManager.Run()
	router := gin.Default()
	router.Use(cors.New(ginHandler.CorsConfig()))
	roomRoute := router.Group("/ws/room")
	{
		roomRoute.GET("/:roomID", ginHandler.RoomConnectHandler(roomManager, db))
	}
	roomNotify := router.Group("/wsServer/notify/room")
	{
		roomNotify.POST("/:id/remove", ginHandler.RoomMemberRemoveHandler(userManager, roomManager, db))
		roomNotify.POST("/:id/join", ginHandler.RoomMemberJoinHandler(userManager, roomManager, db))
		roomNotify.POST("/:id/update", ginHandler.RoomUpdateHandler(userManager, roomManager, db))
	}
	backendUserNotify := router.Group("/wsServer/notify/user")
	{
		backendUserNotify.POST("/:id", ginHandler.BackendUserNotifyHandler(userManager, db))
	}
	UserNotify := router.Group("/wsServer/connection/user")
	{
		UserNotify.GET("/:id", ginHandler.UserNotifyConnectionHandler(userManager, db))
	}

	router.Run(addr)

}
