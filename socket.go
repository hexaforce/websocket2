package main

import (
	"log"
	"net/http"
	"github.com/gorilla/websocket"
)
var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:    4096,
		WriteBufferSize:   4096,
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

// NewSocket new socket.
func NewSocket() *Socket {
	return &Socket{
		broadcast:  make(chan []byte),
		register:   make(chan *Connection),
		unregister: make(chan *Connection),
		connections:    make(map[*Connection]bool),
	}
}

func (h *Socket) open() {
	for {
		select {
		case connection := <-h.register:
			h.connections[connection] = true
		case connection := <-h.unregister:
			h.releaseClient(connection)
		case message := <-h.broadcast:
			for connection := range h.connections {
				select {
				case connection.send <- message:
				default:
					h.releaseClient(connection)
				}
			}
		}
	}
}

// dialUp handles websocket requests from the peer.
func (h *Socket) dialUp(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	connection := &Connection{socket: h, conn: conn, send: make(chan []byte, 256)}
	connection.socket.register <- connection

	// Allow collection of memory referenced by the caller by doing all work in new goroutines.
	go connection.writePump()
	go connection.readPump()

}

func (h *Socket) releaseConnection(connection *Connection){
	if _, ok := h.connections[connection]; ok {
		close(connection.send)
		delete(h.connections, connection)
	}
}
