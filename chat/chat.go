package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"io"
	"net/http"
	"sync"
	//"time"
)

type Server struct {
	connections []*websocket.Conn
	lock        sync.Mutex
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) AddConnection(ws *websocket.Conn) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.connections = append(s.connections, ws)
}

func (s *Server) RemoveConnection(ws *websocket.Conn) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for i, conn := range s.connections {
		if conn == ws {
			s.connections = append(s.connections[:i], s.connections[i+1:]...)
			break
		}
	}
}

func (s *Server) handleWS(ws *websocket.Conn) {
	fmt.Println("new incoming connection from client", ws.RemoteAddr())
	s.AddConnection(ws)
	s.readLoop(ws)
	s.RemoveConnection(ws)
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
	fmt.Println("Connection closed:", ws.RemoteAddr())
}

func (s *Server) broadcast(b []byte, sender *websocket.Conn) {
	for _, ws := range s.connections {
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
