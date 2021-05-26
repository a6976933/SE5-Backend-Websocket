package main

import (
  "github.com/gin-gonic/gin"
  "github.com/gin-contrib/cors"
  "se5-back-websocket/ginHandler"
  "se5-back-websocket/wsHandler"
)

func main() {
  const addr = "127.0.0.1:8000"
  roomManager := wsHandler.NewRoomManager()
  router := gin.Default()
  router.Use(cors.New(ginHandler.CorsConfig()))
  roomRoute := router.Group("/ws/room")
  {
    roomRoute.GET("/:roomID", ginHandler.WsPing(roomManager))
  }

  router.Run(addr)

}
