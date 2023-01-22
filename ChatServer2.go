package main

import (
    "fmt"
    "net/http"
    "reflect"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
  ReadBufferSize:  1024,
  WriteBufferSize: 1024,
}

var chatConns map[string]*websocket.Conn

func messageReadWriter(wsConn *websocket.Conn) {
  for {
    // Read message from browser
    msgType, msg, err := wsConn.ReadMessage()
    if err != nil {
      return
    }

    // Print the message to the console
    fmt.Printf("%s sent: %s\n", wsConn.RemoteAddr(), string(msg))

    // Write message back to browser
    if err = wsConn.WriteMessage(msgType, msg); err != nil {
      return
    }
  }
}

func main() {
  http.HandleFunc("/echo", func(resp http.ResponseWriter, req *http.Request) {
    wsConn, _ := upgrader.Upgrade(resp, req, nil) // error ignored for sake of simplicity
    fmt.Println(reflect.TypeOf(wsConn))
    messageReadWriter(wsConn)
  })

  http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
    http.ServeFile(resp, req, "WebClient2.html")
  })

  http.ListenAndServe(":8080", nil)
}