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

func Serve(
	handlerFunc func(w io.Writer, r *request.Request) *HandlerError,
	port int,
) (*Server, error) {
	portStr := strconv.Itoa(port)
	listener, err := net.Listen("tcp", "127.0.0.1:"+portStr)
	if err != nil {
		return nil, err
	}
	s := &Server{
		state:    stateListening,
		listener: listener,
		handler:  handlerFunc,
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

type HandlerError struct {
	StatusCode int
	Message    string
}

func (he *HandlerError) Error() string {
	return fmt.Sprintf("error: status code: %d, message: %q", he.StatusCode, he.Message)
}

func handleError(w io.Writer, err *HandlerError) {
	h := response.GetDefaultHeaders(0)
	response.WriteStatusLine(w, response.StatusCode(err.StatusCode))
	if err := response.WriteHeaders(w, h); err != nil {
		fmt.Printf("error: %v\n", err)
	}
	w.Write([]byte(err.Message))
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Println("error parsing request:", err)
		return
	}
	buf := bytes.Buffer{}
	handlerErr := s.handler(&buf, req)
	if handlerErr != nil {
		handleError(conn, handlerErr)
		return
	}
	response.WriteStatusLine(conn, response.StatusOK)
	headers := response.GetDefaultHeaders(0)
	if err := response.WriteHeaders(conn, headers); err != nil {
		fmt.Printf("error: %v\n", err)
	}
	conn.Write(buf.Bytes())
}
