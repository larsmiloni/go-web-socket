package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"io"
	"net/http"
)

type Server struct {
	connections map[*websocket.Conn]bool
}

func NewServer() *Server {
	return &Server{
		connections: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) handleWS(ws *websocket.Conn) {
	fmt.Println("new incoming connection from client", ws.RemoteAddr())
	s.connections[ws] = true
	s.readLoop(ws)
}

func (s *Server) readLoop(ws *websocket.Conn) {
	var msg []byte // Use a byte slice to handle binary data
	for {
		err := websocket.Message.Receive(ws, &msg)
		if err != nil {
			if err == io.EOF {
				break // Connection closed normally
			}
			fmt.Println("Read error:", err)
			continue
		}

		s.broadcast(msg, ws)
	}
	delete(s.connections, ws)
	fmt.Println("Connection closed:", ws.RemoteAddr())
}

func (s *Server) broadcast(b []byte, sender *websocket.Conn) {
	for ws := range s.connections {
		if ws != sender {
			go func(ws *websocket.Conn) {
				err := websocket.Message.Send(ws, b)
				if err != nil {
					fmt.Println("Write error:", err)
				}
			}(ws)
		}
	}
}

func main() {
	server := NewServer()
	http.Handle("/ws", websocket.Handler(server.handleWS))
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		fmt.Println("Server error:", err)
	}
}
