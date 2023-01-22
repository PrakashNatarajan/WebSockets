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

type Message struct {
  Sender string
  Recipient string
  Text string
}

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

func jsonMsgReadWriter(wsConn *websocket.Conn) {
  for {
    msg := Message{}
    // Read message from browser
    err := wsConn.ReadJSON(&msg)
    if err != nil {
      fmt.Println("Error reading json.", err)
      return
    }

    // Print the message to the console
    fmt.Printf("Got message: %#v\n", msg)

    // Write message back to browser
    if err = wsConn.WriteJSON(msg); err != nil {
      fmt.Println(err)
      return
    }
  }
}

func echoHandler(resp http.ResponseWriter, req *http.Request) {
  wsConn, _ := upgrader.Upgrade(resp, req, nil) // error ignored for sake of simplicity
  fmt.Println(reflect.TypeOf(wsConn))
  messageReadWriter(wsConn)
}

func chatHandler(resp http.ResponseWriter, req *http.Request) {
  wsConn, _ := upgrader.Upgrade(resp, req, nil) // error ignored for sake of simplicity
  fmt.Println(reflect.TypeOf(wsConn))
  jsonMsgReadWriter(wsConn)
}

func homePage(resp http.ResponseWriter, req *http.Request) {
  http.ServeFile(resp, req, "WebClient2.html")
}

func main() {
  http.HandleFunc("/echo", echoHandler)
  http.HandleFunc("/chat", chatHandler)
  http.HandleFunc("/", homePage)
  http.ListenAndServe(":8080", nil)
}