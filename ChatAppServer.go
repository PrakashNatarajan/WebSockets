package main

import (
    "golang.org/x/net/websocket"
    "fmt"
    "log"
    "net/http"
)

    //var receive string
    type Message struct {
        Sender string `json:"sender"`
        Recipient string `json:"recipient"`
        Text string `json:"text"`
    }

    type UserData struct {
        UserName string `json:"user"`
    }

    type ReplyData struct {
        Code int `json:"code"`
        Status string `json:"status"`
        Message string `json:"message"`
    }

    var loggedinClients = map[string]*websocket.Conn {}
    var chatClients = map[string]*websocket.Conn {}

/*
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(req *http.Request) bool {
        return true
    },
}
*/

func LoginUser(regws *websocket.Conn) {
    var err error

    for {
        
        usrData := &UserData{}

        if err = websocket.JSON.Receive(regws, &usrData); err != nil {
            fmt.Println("Can't receive for register")
            break
        }

        loggedinClients[usrData.UserName] = regws
        
        fmt.Println("\n")
        fmt.Println("Received from client: " + usrData.UserName)
        fmt.Println(usrData)
        fmt.Println(loggedinClients)
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
        
        msgData := &Message{}

        if err = websocket.JSON.Receive(chatws, &msgData); err != nil {
            fmt.Println("Can't receive for chat")
            break
        }
        if client := chatClients[msgData.Sender]; client == nil {
            chatClients[msgData.Sender] = chatws
        }

        toUserws := chatClients[msgData.Recipient]
        if toUserws == nil {
            toUserws = loggedinClients[msgData.Recipient]
        }
        
        fmt.Println("\n")
        fmt.Println(toUserws)
        fmt.Println("Received from client: " + msgData.Sender)
        fmt.Println("Received message: " + msgData.Text)
        fmt.Println("Sending to client: " + msgData.Recipient)
        fmt.Println(msgData)
        fmt.Println("\n")

        if err = websocket.JSON.Send(toUserws, &msgData); err != nil {
            fmt.Println("Can't send to client")
            break
        }

        succReply := ReplyData{Code: 200, Status: "Success", Message: "Sent"}
        if err = websocket.JSON.Send(chatws, &succReply); err != nil {
            fmt.Println("Can't send to user")
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
    http.Handle("/login", websocket.Handler(LoginUser))
    http.Handle("/", websocket.Handler(HomeTest))
    fmt.Println("Started the WebSocketServer with port no: 1234")
    fmt.Println("Use CTRL + C to shutdown")
    if err := http.ListenAndServe(":1234", nil); err != nil {
        log.Fatal("ListenAndServe:", err)
    }
    //if err := http.ListenAndServeTLS(":8000", "cert.pem", "cert.key", nil); err != nil {
    //    panic("ListenAndServeTLS Error: " + err.Error())
    //}
}
