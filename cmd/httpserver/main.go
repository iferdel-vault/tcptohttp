package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
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
	target := r.RequestLine.RequestTarget
	switch {
	case strings.HasPrefix(target, "/httpbin"):
		proxyHandler(w, r)
		return
	case target == "/yourproblem":
		handler400(w, r)
		return
	case target == "/myproblem":
		handler500(w, r)
		return
	default:
		handler200(w, r)
		return
	}
}

func handler400(w *response.Writer, _ *request.Request) {
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
}

func handler500(w *response.Writer, _ *request.Request) {
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
}

func handler200(w *response.Writer, _ *request.Request) {
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

func proxyHandler(w *response.Writer, r *request.Request) {
	target := strings.TrimPrefix(r.RequestLine.RequestTarget, "/httpbin/")
	url := "https://httpbin.org/" + target
	fmt.Println("Proxying to", url)
	resp, err := http.Get(url)
	if err != nil {
		handler500(w, r)
		return
	}
	defer resp.Body.Close()

	w.WriteStatusLine(response.StatusOK)
	h := response.GetDefaultHeaders(0)
	h.Override("Transfer-Encoding", "chunked")
	h.Remove("Content-Length")
	h.Set("Trailer", "X-Content-SHA256, X-Content-Length")
	w.WriteHeaders(h)

	const maxChunkSize = 1024
	buffer := make([]byte, maxChunkSize)
	b := []byte{}
	totalLen := 0
	for {
		n, err := resp.Body.Read(buffer)
		fmt.Println("Read", n, "bytes")
		if n > 0 {
			_, err := w.WriteChunkedBody(buffer[:n])
			if err != nil {
				fmt.Println("Error writing chunked body:", err)
				break
			}
			b = append(b, buffer[:n]...)
			totalLen += n
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading response body:", err)
			break
		}
	}

	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Println("Error writing chunked body done:", err)
	}

	th := response.GetDefaultHeaders(0)
	th.Remove("Content-Length")
	checksum := sha256.Sum256(b)
	hashString := hex.EncodeToString(checksum[:])
	th.Set("X-Content-Sha256", hashString)
	th.Set("X-Content-Length", strconv.Itoa(totalLen))
	err = w.WriteTrailers(th)
	if err != nil {
		fmt.Println("Error writing trailers:", err)
	}
}
