// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//Source ==> https://www.thepolyglotdeveloper.com/2016/12/create-real-time-chat-app-golang-angular-2-websockets/

package main

import (
	"flag"
	"log"
	"net/http"
    "fmt"
)

var addr = flag.String("addr", ":12345", "http service address")

func serveHome(res http.ResponseWriter, req *http.Request) {
    log.Println(req.URL)
    if req.URL.Path != "/" {
        http.Error(res, "Not found", http.StatusNotFound)
        return
    }
    if req.Method != "GET" {
        http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    http.ServeFile(res, req, "home.html")
}

func main() {
    fmt.Println("Starting application...")
    fmt.Println("CTRL+C to shutdown the server")
    flag.Parse()
    manager := newClientManager()
    database := newDBManager()
    manager.database = database
    go manager.start()
    http.HandleFunc("/", serveHome)
    http.HandleFunc("/chat", func(res http.ResponseWriter, req *http.Request) {
        serveWs(manager, res, req)
    })
    err := http.ListenAndServe(*addr, nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
