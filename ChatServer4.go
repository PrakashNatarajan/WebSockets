package main

import (
    "fmt"
    "io"
    "log"
    "net/http"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool { return true },
}

func websocketUpgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return ws, err
    }
    return ws, nil
}

func websocketReader(conn *websocket.Conn) {
    for {
        messageType, p, err := conn.ReadMessage()
        if err != nil {
            log.Println(err)
            return
        }

        fmt.Println(string(p))

        if err := conn.WriteMessage(messageType, p); err != nil {
            log.Println(err)
            return
        }
    }
}

func websocketWriter(conn *websocket.Conn) {
    for {
        fmt.Println("Sending")
        messageType, r, err := conn.NextReader()
        if err != nil {
            fmt.Println(err)
            return
        }
        w, err := conn.NextWriter(messageType)
        if err != nil {
            fmt.Println(err)
            return
        }
        if _, err := io.Copy(w, r); err != nil {
            fmt.Println(err)
            return
        }
        if err := w.Close(); err != nil {
            fmt.Println(err)
            return
        }
    }
}

func serveWs(w http.ResponseWriter, r *http.Request) {
    ws, err := websocketUpgrade(w, r)
    if err != nil {
        fmt.Fprintf(w, "%+V\n", err)
    }
    go websocketWriter(ws)
    websocketReader(ws)
}

func setupRoutes() {
    http.HandleFunc("/ws", serveWs)
}

func main() {
    fmt.Println("Distributed Chat App v0.01")
    setupRoutes()
    http.ListenAndServe(":8080", nil)
}