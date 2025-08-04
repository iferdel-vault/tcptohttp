package server

import (
	"fmt"
	"net"
	"strconv"
	"sync/atomic"
)

type Server struct {
	state    serverState
	listener net.Listener
	isClosed atomic.Bool
}

type serverState int

const (
	stateListening serverState = iota
)

func Serve(port int) (*Server, error) {
	portStr := strconv.Itoa(port)
	listener, err := net.Listen("tcp", "127.0.0.1:"+portStr)
	if err != nil {
		return nil, err
	}
	s := &Server{
		state:    stateListening,
		listener: listener,
	}
	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	s.isClosed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.isClosed.Load() {
				return
			}
			fmt.Println("error in listen", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello World!"))
	conn.Close()
}
