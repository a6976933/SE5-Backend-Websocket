package ginHandler

import (
  "github.com/gin-gonic/gin"
  "github.com/gorilla/websocket"
  "se5-back-websocket/wsHandler"
  //"golang.org/x/crypto/ssh"
  "log"
  "encoding/json"
  //"io/ioutil"
  "strconv"
  "time"
  "github.com/dgrijalva/jwt-go"
)

//openssl genrsa -out key.pem 2048
//openssl rsa -in key.pem -pubout -out key.pem.pub

const (
  //SECRETKEYPATH = "key.pem"
  MAX_AGE = 300
)

type jwtRet struct {
  JWT string `json:"token"`
}

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
    /*
    secretKeyFile, err := ioutil.ReadFile(SECRETKEYPATH)
    if err != nil {
      log.Println(err)
    }
    */
    key := []byte(wsHandler.KEY)


    //key, err := ssh.ParseRawPrivateKey(secretKeyFile)

    if err != nil {
      log.Println(err)
    }


    customClaim := &wsHandler.JWTClaim{
      UserID: initInfo.UserID,
      Username: initInfo.Username,
      //Pubkey: pubkey,
      StandardClaims: jwt.StandardClaims{
        ExpiresAt: time.Now().Add(time.Duration(MAX_AGE)*time.Second).Unix(),
        Issuer:initInfo.Username,
      },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaim)
    tokenString, err := token.SignedString(key)
    log.Println(tokenString)
    jwtObj := &jwtRet{ JWT: tokenString}
    sendJWT, _ := json.Marshal(jwtObj)
    wsH.Conn.WriteMessage(websocket.TextMessage, sendJWT)
    if err != nil {
      log.Println(err)
    }
    wsH.Username = initInfo.Username
    wsH.UserID = initInfo.UserID
    wsH.Room = servingRoom
    wsH.Register()
    wsH.FetchMessage()
    go wsH.ReadPump()
    go wsH.WritePump()
  }
  return gin.HandlerFunc(fn)
}
