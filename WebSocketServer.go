package main

import (
    "golang.org/x/net/websocket"
    "fmt"
    "log"
    "net/http"
)

    //var reply string
    type socketData struct {
        FrmUsr string `json:"frmusr"`
        ToUsr string `json:"tousr"`
        Msg string `json:"msg"`
    }

func Echo(ws *websocket.Conn) {
    var err error

    for {
        
        reply := &socketData{}

        if err = websocket.JSON.Receive(ws, &reply); err != nil {
            fmt.Println("Can't receive")
            break
        }

        fmt.Println("Received back from client: " + reply.Msg)
        fmt.Println(reply)

        msg := "Received:  " + reply.Msg
        fmt.Println("Sending to client: " + msg)

        if err = websocket.JSON.Send(ws, &reply); err != nil {
            fmt.Println("Can't send")
            break
        }
    }
}

func main() {
    http.Handle("/", websocket.Handler(Echo))
    fmt.Println("Started the WebSocketServer with port no: 1234")
    fmt.Println("Use CTRL + C to shutdown")
    if err := http.ListenAndServe(":1234", nil); err != nil {
        log.Fatal("ListenAndServe:", err)
    }
}
