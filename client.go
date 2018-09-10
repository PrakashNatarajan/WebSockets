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
	send chan []byte
}


type Message struct {
    Sender    string `json:"sender,omitempty"`
    Recipient string `json:"recipient,omitempty"`
    Content   string `json:"content,omitempty"`
}


func (manager *ClientManager) send(message []byte, ignore *Client) {
    for conn := range manager.clients {
        if conn != ignore {
            conn.send <- message
        }
    }
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

    for {
        _, message, err := client.socket.ReadMessage()
        if err != nil {
            manager.unregister <- client
            client.socket.Close()
            break
        }
        jsonMessage, _ := json.Marshal(&Message{Sender: client.id, Content: string(message)})
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

    for {
        select {
        case message, ok := <-client.send:
            if !ok {
                client.socket.WriteMessage(websocket.CloseMessage, []byte{})
                return
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
    client := &Client{id: guid.String(), socket: conn, send: make(chan []byte)}

    manager.register <- client

    go client.read(manager)
    go client.write(manager)
}