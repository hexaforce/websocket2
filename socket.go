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
		broadcast:   make(chan []byte),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
		connections: make(map[*Connection]bool),
	}
}

func (s *Socket) open() {
	for {
		select {
		case connection := <-s.register:
			s.connections[connection] = true
		case connection := <-s.unregister:
			s.releaseConnection(connection)
		case message := <-s.broadcast:
			for connection := range s.connections {
				select {
				case connection.send <- message:
				default:
					s.releaseConnection(connection)
				}
			}
		}
	}
}

// dialUp handles websocket requests from the peer.
func (s *Socket) dialUp(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	connection := &Connection{socket: s, wsConn: conn, send: make(chan []byte, 256)}
	connection.socket.register <- connection

	// Allow collection of memory referenced by the caller by doing all work in new goroutines.
	go connection.writePump()
	go connection.readPump()

}

func (s *Socket) releaseConnection(connection *Connection) {
	if _, ok := s.connections[connection]; ok {
		close(connection.send)
		delete(s.connections, connection)
	}
}
