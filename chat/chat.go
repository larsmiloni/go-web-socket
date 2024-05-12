package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"io"
	"net/http"
	"sync"
)

type Server struct {
	connections map[*websocket.Conn]bool
	mutex       sync.Mutex
}

func NewServer() *Server {
	return &Server{
		connections: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) handleWS(ws *websocket.Conn) {
	fmt.Println("new incoming connection from client", ws.RemoteAddr())

	s.mutex.Lock()
	s.connections[ws] = true
	s.mutex.Unlock()

	s.readLoop(ws)
}

func (s *Server) readLoop(ws *websocket.Conn) {
	buff := make([]byte, 1024)
	for {
		n, err := ws.Read(buff)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("read error:", err)
			continue
		}
		msg := buff[:n]
		s.broadcast(msg, ws)
	}

	s.mutex.Lock()
	delete(s.connections, ws)
	s.mutex.Unlock()

	fmt.Println("Connection closed:", ws.RemoteAddr())
}

func (s *Server) broadcast(b []byte, sender *websocket.Conn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for ws := range s.connections {
		if ws != sender {
			go func(ws *websocket.Conn) {
				if _, err := ws.Write(b); err != nil {
					fmt.Println("write error:", err)
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
		return
	}
}
