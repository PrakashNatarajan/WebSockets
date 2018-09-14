// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
    "fmt"
    "encoding/json"
)

// ClientManager maintains the set of active clients and broadcasts messages to the
// clients.
type ClientManager struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

    // Register requests from the clients.
    regClients map[string]*Client

    // Connected Database.
    database *DBManager

    // Registered client Ids.
    ClientIds []string
}

func newClientManager() *ClientManager {
	return &ClientManager{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
        regClients: make(map[string]*Client),
        //database: make(*DBManager),
        ClientIds: make([]string, 100),
	}
}

func (manager *ClientManager) send(message []byte, ignore *Client) {
    for conn := range manager.clients {
        if conn != ignore {
            conn.sendMsg <- message
        }
    }
}

func (manager *ClientManager) sendClientIds(ignore *Client) {
    for conn := range manager.clients {
        var clientIds []string
        fmt.Println("ClientIds: ", manager.ClientIds)
        clientIds = removeDuplicates(manager.ClientIds)
        fmt.Println("Uniq Ids: ", clientIds)
        jsonMessage, _ := json.Marshal(clientIds)
        conn.sendIds <- jsonMessage
    }
}

func (manager *ClientManager) start() {
    for {
        select {
        case conn := <-manager.register:
            manager.clients[conn] = true
            manager.regClients[conn.id] = conn
            manager.ClientIds = append(manager.ClientIds, conn.id)
            //jsonMessage, _ := json.Marshal(&Message{Content: "Id: " + conn.id})
            manager.sendClientIds(conn)
        case conn := <-manager.unregister:
            if _, ok := manager.clients[conn]; ok {
                close(conn.sendMsg)
                delete(manager.clients, conn)
                jsonMessage, _ := json.Marshal(&Message{Content: "/A socket has disconnected."})
                manager.send(jsonMessage, conn)
            }
        case message := <-manager.broadcast:
            for conn := range manager.clients {
                select {
                case conn.sendMsg <- message:
                default:
                    close(conn.sendMsg)
                    delete(manager.clients, conn)
                }
            }
        }
    }
}


func removeDuplicates(elements []string) []string {
    // Use map to record duplicates as we find them.
    encountered := map[string]bool{}
    result := []string{}

    for v := range elements {
        if encountered[elements[v]] == true {
            // Do not add duplicate.
        } else {
            // Record this element as an encountered element.
            encountered[elements[v]] = true
            // Append to result slice.
            result = append(result, elements[v])
        }
    }
    // Return the new slice.
    return result
}