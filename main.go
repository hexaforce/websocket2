package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
)

// Connection is a middleman between the websocket connection and the Socket.
type Connection struct {
	socket *Socket
	// The websocket connection.
	wsConn *websocket.Conn
	// Buffered channel of outbound messages.
	send chan []byte
}

// Socket maintains the set of active connections and broadcasts messages to the connections.
type Socket struct {
	// Registered connections.
	connections map[*Connection]bool
	// Inbound messages from the connections.
	broadcast chan []byte
	// Register requests from the connections.
	register chan *Connection
	// Unregister requests from connections.
	unregister chan *Connection
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var addr = flag.String("addr", ":18080", "http service address")

func main() {

	flag.Parse()
	socket := &Socket{
		broadcast:   make(chan []byte),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
		connections: make(map[*Connection]bool),
	}
	go socket.open()

	router := httprouter.New()
	router.GET("/ws", socket.dial)

	handler := cors.Default().Handler(router)
	log.Fatal("ListenAndServe: ", http.ListenAndServe(*addr, handler))

}
