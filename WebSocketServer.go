package main

import (
    "golang.org/x/net/websocket"
    "fmt"
    "log"
    "net/http"
)

    //var reply string
    type ChatData struct {
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
        
        fmt.Println("\n")
        fmt.Println("Received from client: " + usrData.UserName)
        fmt.Println(usrData)
        fmt.Println(registeredClients)
        fmt.Println("Sending back to client: " + usrData.UserName)
        fmt.Println("\n")

        if err = websocket.JSON.Send(regws, &usrData); err != nil {
            fmt.Println("Can't send")
            break
        }
    }
}

func ChatMessage(chatws *websocket.Conn) {
    var err error

    for {
        
        reply := &ChatData{}

        if err = websocket.JSON.Receive(chatws, &reply); err != nil {
            fmt.Println("Can't receive for chat")
            break
        }

        toUserws := registeredClients[reply.ToUsr]

        fmt.Println("\n")
        fmt.Println(toUserws)
        fmt.Println("Received from client: " + reply.FrmUsr)
        fmt.Println("Received message: " + reply.Msg)
        fmt.Println("Sending to client: " + reply.ToUsr)
        fmt.Println(reply)
        fmt.Println("\n")

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
