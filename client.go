// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
    "net/http"
    "time"
    "encoding/json"
    "github.com/gorilla/websocket"
    "github.com/satori/go.uuid"
    "fmt"
)

const (
    // Time allowed to write a message to the peer.
    writeWait = 10 * time.Second

    // Time allowed to read the next pong message from the peer.
    pongWait = 60 * time.Second

    // Send pings to peer with this period. Must be less than pongWait.
    pingPeriod = (pongWait * 9) / 10

    // Maximum message size allowed from peer.
    maxMessageSize = 512
)

var (
    newline = []byte{'\n'}
    space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
    id     string

    // The websocket connection.
    socket *websocket.Conn

    // Buffered channel of outbound messages.
    sendMsg chan []byte

    // Buffered channel of outbound clientids.
    sendIds chan []byte
}

type Message struct {
    Guid      string `json:"guid,omitempty"`
    Sender    string `json:"sender,omitempty"`
    Receiver string `json:"receiver,omitempty"`
    Content   string `json:"content,omitempty"`
}

type ReciContent struct {
    Receiver string `json:"receiver,omitempty"`
    Content   string `json:"content,omitempty"`
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (client *Client) read(manager *ClientManager) {
    defer func() {
        manager.unregister <- client
        client.socket.Close()
    }()

    var recicont ReciContent
    database := manager.database

    for {
        _, message, err := client.socket.ReadMessage()
        if err != nil {
            manager.unregister <- client
            client.socket.Close()
            break
        }
        err = json.Unmarshal(message, &recicont)
        if err != nil {
            fmt.Println("error:", err)
            manager.unregister <- client
            client.socket.Close()
            break
        }
        //fmt.Println("recicont:", recicont.Content, "Receiver: ", recicont.Receiver)
        guid, _ := uuid.NewV4()
        msgobj := Message{Guid: guid.String(), Sender: client.id, Content: recicont.Content, Receiver: recicont.Receiver}
        database.CreateRecord(msgobj.Guid, msgobj.Sender, msgobj.Content, msgobj.Receiver)
        jsonMessage, _ := json.Marshal(&msgobj)
        manager.broadcast <- jsonMessage
    }
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (client *Client) write(manager *ClientManager) {
    defer func() {
        client.socket.Close()
    }()

    var msgcont Message
    var clientIds []string
    database := manager.database
    for {
        select {
        case message, ok := <-client.sendMsg:
            if !ok {
                client.socket.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            //fmt.Println(string(message)) //Converts bytes to string as readable format.
            err := json.Unmarshal(message, &msgcont)
            if err != nil {
                fmt.Println("Write function error:", err)
                manager.unregister <- client
                client.socket.Close()
                break
            }
            fmt.Println("recicont:", msgcont.Content, "Receiver: ", msgcont.Receiver)
            msgstatus := database.GetRecordStatus(msgcont.Guid)
            if msgstatus == "UnSent" {
                sentClient := manager.regClients[msgcont.Sender]
                reciClient := manager.regClients[msgcont.Receiver]
                defer func() {
                    reciClient.socket.Close()
                }()
                reciClient.socket.WriteMessage(websocket.TextMessage, message)
                sentClient.socket.WriteMessage(websocket.TextMessage, message)
                database.UpdateRecord(msgcont.Guid, "Sent")
            } else {
                return
            }
        case message, ok := <-client.sendIds:
            if !ok {
                client.socket.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            //fmt.Println(string(message)) //Converts bytes to string as readable format.
            err := json.Unmarshal(message, &clientIds)
            if err != nil {
                fmt.Println("Write function error:", err)
                manager.unregister <- client
                client.socket.Close()
                break
            }
            client.socket.WriteMessage(websocket.TextMessage, message)
        }
    }
}

// serveWs handles websocket requests from the peer.
func serveWs(manager *ClientManager, res http.ResponseWriter, req *http.Request) {
    conn, error := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(res, req, nil)
    if error != nil {
        http.NotFound(res, req)
        return
    }
    guid, _ := uuid.NewV4()
    client := &Client{id: guid.String(), socket: conn, sendMsg: make(chan []byte), sendIds: make(chan []byte)}

    manager.register <- client

    go client.read(manager)
    go client.write(manager)
}