package ginHandler

import (
  "github.com/gin-gonic/gin"
  "se5-back-websocket/wsHandler"
  "log"
  "encoding/json"
  "strconv"
)

func WsPing(rm *wsHandler.RoomManager) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    var roomID int
    var err error
    var servingRoom *wsHandler.RoomMsgManager
    var initInfo = new(wsHandler.RecvMsg)
    wsH := wsHandler.NewWsHandler()
    /*
    if reflect.TypeOf(c.Param("roomID")).String() != "string"{
      c.String("404", "Page Not Found")
    }*/
    if roomID, err = strconv.Atoi(c.Param("roomID")); err != nil {
      log.Println(err)
      c.String(404, "Page Not Found")
      return
    }
    if !rm.IsRoomExist(roomID) {
      servingRoom = rm.CreateRoom(roomID)
    } else {
      servingRoom = rm.LiveRoomList[roomID]
    }
    go servingRoom.Run()
    err = wsH.InitWebsocketConn(c)
    if err != nil {
      log.Println(err)
      c.String(400, "Websocket error")
      return
    } else {
      log.Println("Someone connect to the room ID: ", roomID)
    }
    _, message, err := wsH.Conn.ReadMessage()
    if err != nil {
      log.Println(err)
      //c.String(400, "Data is wrong")
      return
    }
    err = json.Unmarshal(message, &initInfo)
    if err != nil {
      log.Println(err)
      //c.String(400, "Data is wrong")
      return
    } else {
      log.Println("Has got the user information, \n name: ", initInfo.Username)
    }
    wsH.Username = initInfo.Username
    wsH.UserID = initInfo.UserID
    wsH.Room = servingRoom
    wsH.Register()
    go wsH.ReadPump()
    go wsH.WritePump()
  }
  return gin.HandlerFunc(fn)
}
