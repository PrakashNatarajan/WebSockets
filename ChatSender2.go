//Source ==> https://tutorialedge.net/projects/chat-system-in-go-and-react/part-4-handling-multiple-clients/

package main

import (
    "fmt"
    "log"
    //"sync"
    "net/http"

    "github.com/gorilla/websocket"
    amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
    ID   string
    Name string
    Conn *websocket.Conn
    Pool *Pool
}

type Pool struct {
    Register   chan *Client
    Unregister chan *Client
    Clients    map[*Client]bool
}

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(req *http.Request) bool { return true },
}

func NewPool() *Pool {
    return &Pool{
        Register:   make(chan *Client),
        Unregister: make(chan *Client),
        Clients:    make(map[*Client]bool),
    }
}

type Message struct {
    Type int    `json:"type"`
    Body string `json:"body"`
}

func (c *Client) Read() {
    defer func() {
        c.Pool.Unregister <- c
        c.Conn.Close()
    }()

    for {
        messageType, p, err := c.Conn.ReadMessage()
        if err != nil {
            log.Println(err)
            return
        }
        //[]byte to convert as string string([]byte) []byte is not readable
        message := Message{Type: messageType, Body: string(p)}
        //c.Pool.Broadcast <- message
        fmt.Printf("Message Received: %+v\n", message)
    }
}

func (pool *Pool) ManageClientConns() {
    for {
        select {
        case client := <-pool.Register:
            pool.Clients[client] = true
            fmt.Println("Size of Connection Pool: ", len(pool.Clients))
            for client, _ := range pool.Clients {
                fmt.Println(client)
                client.Conn.WriteJSON(Message{Type: 1, Body: "New User Joined..."})
            }
            break
        case client := <-pool.Unregister:
            delete(pool.Clients, client)
            fmt.Println("Size of Connection Pool: ", len(pool.Clients))
            for client, _ := range pool.Clients {
                client.Conn.WriteJSON(Message{Type: 1, Body: "User Disconnected..."})
            }
            break            
        }
    }
}

func connUpgrade(res http.ResponseWriter, req *http.Request) (*websocket.Conn, error) {
    conn, err := upgrader.Upgrade(res, req, nil)
    if err != nil {
        log.Println(err)
        return nil, err
    }

    return conn, nil
}

func failOnError(err error, msg string) {
  if err != nil {
    log.Panicf("%s: %s", msg, err)
  }
}

func (pool *Pool)EventQueueConn() (conn *amqp.Connection, channel *amqp.Channel, queue amqp.Queue) {
  conn, err := amqp.Dial("amqp://MsgQueAdmin:MyQueueS123@172.17.0.2:5672/")
  failOnError(err, "Failed to connect to RabbitMQ")

  channel, err = conn.Channel()
  failOnError(err, "Failed to open a channel")

  queue, err = channel.QueueDeclare(
    "TestQueue", // name
    false,   // durable
    false,   // delete when unused
    false,   // exclusive
    false,   // no-wait
    nil,     // arguments
  )
  failOnError(err, "Failed to declare a queue")
  return conn, channel, queue
}

func (pool *Pool)ReceiveQueueMsgs() {
  conn, channel, queue := pool.EventQueueConn()
  defer conn.Close()
  defer channel.Close()
  messages, err := channel.Consume(
    queue.Name, // queue
    "",     // consumer
    true,   // auto-ack
    false,  // exclusive
    false,  // no-local
    false,  // no-wait
    nil,    // args
  )
  failOnError(err, "Failed to receive messages from queue")
  pool.ReceiveSendMsgs(messages)

}

func (pool *Pool)ReceiveSendMsgs(messages <-chan amqp.Delivery) {
  for quMsg := range messages {
    log.Printf("Received a message: %s", quMsg.Body)
    fmt.Println("Sending message to all clients in Pool")
    for client, _ := range pool.Clients {
        msgData := Message{Type: 1, Body: string(quMsg.Body)}
        if err := client.Conn.WriteJSON(msgData); err != nil {
            fmt.Println(err)
            return
        }
    }
  }
}

func serveWs(pool *Pool, res http.ResponseWriter, req *http.Request) {
    fmt.Println("WebSocket Endpoint Hit")
    conn, err := connUpgrade(res, req)
    if err != nil {
        fmt.Fprintf(res, "%+v\n", err)
    }

    client := &Client{
        Conn: conn,
        Pool: pool,
    }

    pool.Register <- client
    client.Read()
}

func setupRoutes() {
  pool := NewPool()
  go pool.ManageClientConns()

  go pool.ReceiveQueueMsgs()
  
  http.HandleFunc("/ws", func(res http.ResponseWriter, req *http.Request) {
    serveWs(pool, res, req)
  })

}

func main() {
    fmt.Println("Distributed Chat App v0.01")
    setupRoutes()
    http.ListenAndServe(":8080", nil)
}
