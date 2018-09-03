package main

import (
    "golang.org/x/net/websocket"
    "fmt"
    "log"
    "net/http"
)

    //var reply string
    type SocketData struct {
        FrmUsr string `json:"frmusr"`
        ToUsr string `json:"tousr"`
        Msg string `json:"msg"`
    }

    type UserData struct {
        UserName string `json:"user"`
    }

    var registeredClients = map[string]*websocket.Conn {}

func RegisterUser(regws *websocket.Conn) {
    var err error

    for {
        
        usrData := &UserData{}

        if err = websocket.JSON.Receive(regws, &usrData); err != nil {
            fmt.Println("Can't receive for register")
            break
        }

        registeredClients[usrData.UserName] = regws
        
        fmt.Println("Received back from client: " + usrData.UserName)
        fmt.Println(usrData)
        fmt.Println(registeredClients)

        usr := "Received:  " + usrData.UserName
        fmt.Println("Sending to client: " + usr)

        if err = websocket.JSON.Send(regws, &usrData); err != nil {
            fmt.Println("Can't send")
            break
        }
    }
}

func ChatMessage(ws *websocket.Conn) {
    var err error

    for {
        
        reply := &SocketData{}

        if err = websocket.JSON.Receive(ws, &reply); err != nil {
            fmt.Println("Can't receive for chat")
            break
        }

        toUserws := registeredClients[reply.ToUsr]

        fmt.Println("Received back from client: " + reply.FrmUsr)
        fmt.Println(reply)
        fmt.Println(toUserws)

        msg := "Received:  " + reply.Msg
        fmt.Println("Sending to client: " + msg)

        if err = websocket.JSON.Send(toUserws, &reply); err != nil {
            fmt.Println("Can't send")
            break
        }
    }
}

func HomeTest(ws *websocket.Conn) {
    var err error

    for {

        if err = websocket.Message.Send(ws, "HomePageTest"); err != nil {
            fmt.Println("Can't send")
            break
        }
    }
}

func main() {
    http.Handle("/chat", websocket.Handler(ChatMessage))
    http.Handle("/reg", websocket.Handler(RegisterUser))
    http.Handle("/", websocket.Handler(HomeTest))
    fmt.Println("Started the WebSocketServer with port no: 1234")
    fmt.Println("Use CTRL + C to shutdown")
    if err := http.ListenAndServe(":1234", nil); err != nil {
        log.Fatal("ListenAndServe:", err)
    }
}
