package main

type Chat struct {
    //Clients Connected
    connections map[*Connection]bool

    //Messages to the chat from clients
    incoming chan []byte

    //Connection Requests
    join chan *Connection

    //Disconnection Requests
    leave chan *Connection
}

//Initialize a chat room with nil values
var chat = Chat{
    connections: make(map[*Connection]bool),
    incoming:    make(chan []byte),
    join:        make(chan *Connection),
    leave:       make(chan *Connection),
}

//Dictates the behavior of the chat room while running
func (c *Chat) run() {
    for {
        select {
        case conn := <-c.join:
            c.connections[conn] = true //when a new member connects, add them to the map of connections
        case conn := <-c.leave:
            if _, ok := c.connections[conn]; ok {
                 delete(c.connections, conn)
                 close(conn.send)
            }                          //when a member disconnects, check to see if they are in the map, and if so delete them from it
        case msg := <-c.incoming:
            for conn := range c.connections {
                select {
                case conn.send <- msg:
                default:
                    close(conn.send)
                    delete(chat.connections, conn)
                }
            }                          //
        }

    }
}
