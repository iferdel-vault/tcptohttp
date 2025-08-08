package server

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/iferdel-vault/tcptohttp/internal/request"
	"github.com/iferdel-vault/tcptohttp/internal/response"
)

type Handler func(w io.Writer, r *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (he HandlerError) Write(w io.Writer) {
	response.WriteStatusLine(w, he.StatusCode)
	messageBytes := []byte(he.Message)
	headers := response.GetDefaultHeaders(len(messageBytes))
	if err := response.WriteHeaders(w, headers); err != nil {
		fmt.Printf("error: %v\n", err)
	}
	w.Write(messageBytes)
}

// Server is an HTTP 1.1 server
type Server struct {
	state    serverState
	listener net.Listener
	isClosed atomic.Bool
	handler  func(w io.Writer, r *request.Request) *HandlerError
}

type serverState int

const (
	stateListening serverState = iota
)

func Serve(port int, handler Handler) (*Server, error) {
	portStr := strconv.Itoa(port)
	listener, err := net.Listen("tcp", "127.0.0.1:"+portStr)
	if err != nil {
		return nil, err
	}
	s := &Server{
		state:    stateListening,
		listener: listener,
		handler:  handler,
	}
	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	s.isClosed.Store(true)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
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
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    err.Error(),
		}
		hErr.Write(conn)
		return
	}
	buf := bytes.NewBuffer([]byte{})
	hErr := s.handler(buf, req)
	if hErr != nil {
		hErr.Write(conn)
		return
	}
	b := buf.Bytes()
	response.WriteStatusLine(conn, response.StatusOK)
	headers := response.GetDefaultHeaders(len(b))
	if err := response.WriteHeaders(conn, headers); err != nil {
		hErr := &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    err.Error(),
		}
		hErr.Write(conn)
		return
	}
	conn.Write(b)
	return
}
