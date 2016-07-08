package main

import (
    "fmt"
    "flag"
    "log"
    "net/http"
    "html/template"
    "io/ioutil"

    "github.com/Joker/jade"
)

var chatHost = flag.String("host", ":8080", "http service address")
var jadeFile = flag.String("jade", "template.jade", "chatroom jade template")

func homeHandler(resp http.ResponseWriter, req *http.Request) {

    if req.URL.Path != "/" {
         http.Error(resp, "File Not Found", 404)
         return
    }
    if req.Method != "GET" {
         http.Error(resp, "Unauthorized Method", 405)
         return
    }

    buf, err := ioutil.ReadFile(*jadeFile)
    if err != nil {
        fmt.Printf("\nReadFile error: %v", err)
        return
    }
    jadeTemp, err := jade.Parse("jade_tp", string(buf))
    if err != nil {
        fmt.Printf("\nParse error: %v", err)
        return
    }

    goTemp, err := template.New("html").Parse(jadeTemp)
    if err != nil {
        fmt.Printf("\nTemplate parse error: %v", err)
        return
    }

    //fmt.Printf(jadeTemp)
    resp.Header().Set("Content-Type", "text/html; charset=utf-8")
    err = goTemp.Execute(resp, req.Host)
    if err != nil {
        fmt.Printf("\nExecute error: %v", err)
    }
}

func main() {
    flag.Parse()
    go chat.run()
    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/ws", wsHandler)
    err := http.ListenAndServe(*chatHost, nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
