package main

import (
  "github.com/gin-gonic/gin"
  "github.com/gin-contrib/cors"
  "se5-back-websocket/ginHandler"
  "se5-back-websocket/wsHandler"
  "gorm.io/gorm"
  "log"
  //"strconv"
  //"time"
  "gorm.io/driver/sqlite"
)


func initDatabase(dsn string) (*gorm.DB, error) {
  db, err := gorm.Open(sqlite.Open(dsn),&gorm.Config{})
  if err != nil {
    log.Println(err)
    return nil, err
  }
  return db, nil
}

func main() {
  const addr = "127.0.0.1:8090"
  const dsn = "/Users/allenwang/se5-back/db.sqlite3"
  db, err := initDatabase(dsn)
  if err != nil {
    return
  }
/*
  for i := 1; i < 10 ; i++ {
    test := wsHandler.Room_Roommessage{ID:i+118,RoomID:1445,MemberID:22,Message:"靠"+strconv.Itoa(i),RecvTime:time.Now()}
    result := db.Create(&test)
    if result.Error != nil {
      log.Println(result.Error)
      return
    }
  }*/
  roomManager := wsHandler.NewRoomManager()
  router := gin.Default()
  router.Use(cors.New(ginHandler.CorsConfig()))
  roomRoute := router.Group("/ws/room")
  {
    roomRoute.GET("/:roomID", ginHandler.WsPing(roomManager, db))
  }
  //roomNotify := router.Group("/wsServer/notify")
  {
    //roomNotify.POST("/room/:id/remove")
    //roomNotify.POST("/room/:id/invite")
    //roomNotify.POST("/room/:id/block")
    //roomNotify.POST("/room/:id/priorityChange")
    //roomNotify.POST("/room/:id/join")
  }

  router.Run(addr)

}
