package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/iferdel-vault/tcptohttp/internal/request"
	"github.com/iferdel-vault/tcptohttp/internal/response"
	"github.com/iferdel-vault/tcptohttp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, r *request.Request) {
	switch r.RequestLine.RequestTarget {
	case "/yourproblem":
		w.WriteStatusLine(response.StatusBadRequest)
		body := []byte(`
<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>
`)
		headers := response.GetDefaultHeaders(len(body))
		headers.Override("Content-Type", "text/html")
		w.WriteHeaders(headers)
		w.WriteBody(body)
	case "/myproblem":
		w.WriteStatusLine(response.StatusInternalServerError)
		body := []byte(`
<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>
`)
		headers := response.GetDefaultHeaders(len(body))
		headers.Override("Content-Type", "text/html")
		w.WriteHeaders(headers)
		w.WriteBody(body)
	default:
		w.WriteStatusLine(response.StatusOK)
		body := []byte(`
<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>
`)
		headers := response.GetDefaultHeaders(len(body))
		headers.Override("Content-Type", "text/html")
		w.WriteHeaders(headers)
		w.WriteBody(body)
	}
}
