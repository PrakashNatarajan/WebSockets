package main

import (
    "fmt"
    "net/http"

    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

func main() {
    http.HandleFunc("/echo", func(resp http.ResponseWriter, req *http.Request) {
        conn, _ := upgrader.Upgrade(resp, req, nil) // error ignored for sake of simplicity

        for {
            // Read message from browser
            msgType, msg, err := conn.ReadMessage()
            if err != nil {
                return
            }

            // Print the message to the console
            fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))

            // Write message back to browser
            if err = conn.WriteMessage(msgType, msg); err != nil {
                return
            }
        }
    })

    http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
        http.ServeFile(resp, req, "WebClient2.html")
    })

    http.ListenAndServe(":8080", nil)
}