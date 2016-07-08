package main

import (
    "bytes"
    "log"
    "net/http"
    "time"

    "github.com/gorilla/websocket"
)

const (
    // Wait this long between message send requests
    writeWait = 10 * time.Second

    // The message will stay up this long
    readWait = 60 * time.Second

    // Check for new messages this frequently
    msgPing = (readWait * 9) / 10

    // Maximum message size
    maxSize = 512
)

// For convenience, declare these as vars
var (
     newline = []byte{'\n'}
     space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

// Define what a connection to the chat is
type Connection struct {
     // web socket connection, of course
     ws *websocket.Conn

     // buffered channel for sending messages to the chat
     send chan []byte
}

// function for sending messages to the chat over connections
func (c *Connection) toChat() {
    defer func() {
        chat.leave <- c // send the connection to the chat's leave channel when the loop exits
        c.ws.Close()    // close the connection's websocket
    }()
    // Configure the connection's websocket
    c.ws.SetReadLimit(maxSize)
    c.ws.SetReadDeadline(time.Now().Add(readWait))
    c.ws.SetPongHandler(func(string) error {c.ws.SetReadDeadline(time.Now().Add(readWait)); return nil})
    // Collect messages from the connection's websocket and pass them to the chat
    for {
        _, msg, err := c.ws.ReadMessage() // read connection's websocket
        if err != nil {                   // handle errors
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
                log.Printf("\nUnexpected Close error: %v", err)
            }
            break                         // on error, break out and disconnect
        }
        // trim down the message, cutting off leading and trailing whitespace
        msg = bytes.TrimSpace(bytes.Replace(msg, newline, space, -1))
        // send the message to the chat
        chat.incoming <- msg
    }
}

// Function for writing to a connection's websocket
func (c *Connection) write(mType int, message []byte) error {
     c.ws.SetWriteDeadline(time.Now().Add(writeWait))
     return c.ws.WriteMessage(mType, message)
}

// Read messages from the chat and write to a connection's websocket
func (c *Connection) fromChat() {
    clock := time.NewTicker(msgPing) // Keep track of time
    defer func() {                   // On disconnect, stop clock and close websocket
         clock.Stop()
         c.ws.Close()
    }()
    // Loop indefinitely...
    for {
        select {
        case msg, ok := <-c.send:   // Receive message from connection send channel if available
            if !ok { // if not yet available, close after deadline
                c.write(websocket.CloseMessage, []byte{})
                return
            }
            // message available, wait til deadline to receive...
            c.ws.SetWriteDeadline(time.Now().Add(writeWait))
            w, err := c.ws.NextWriter(websocket.TextMessage)
            if err != nil {
                 return
            }
            w.Write(msg)

            n := len(c.send)
            for i := 0; i < n; i++ {
                 w.Write(newline)
                 w.Write(<-c.send)
            }

            if err := w.Close(); err != nil {
                 return
            }
        case <-clock.C:
            if err := c.write(websocket.PingMessage, []byte{}); err != nil {
                 return
            }
        }
    }
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil{
        log.Println(err)
        return
    }
    conn := &Connection{send: make(chan []byte, 256), ws: ws}
    chat.join <- conn
    go conn.fromChat()
    conn.toChat()
}
