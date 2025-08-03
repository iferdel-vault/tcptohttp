package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/iferdel-vault/tcptohttp/internal/request"
)

const port = "42069"

func main() {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("error setting up tcp listener: %s\n", err.Error())
	}
	defer listener.Close()

	fmt.Println("Reading data from listener on port", port)
	fmt.Println("=====================================")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("error generating connection to listener: %s\n", err.Error())
		}
		fmt.Println("connection has been accepted from", conn.RemoteAddr())

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("error request from reader: %s\n", err.Error())
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for key, value := range req.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}

		fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
	}

}

func getLinesChannel(f io.ReadCloser) <-chan string {
	c := make(chan string)
	b := make([]byte, 8, 8)
	var line string
	go func(c chan<- string) {
		for {
			n, err := f.Read(b)
			if err != nil {
				if line != "" {
					c <- fmt.Sprintf("%s", line)
					line = ""
				}
				if errors.Is(err, io.EOF) {
					close(c)
					break
				}
				log.Fatalf("error reading file: %v", err)
				break
			}
			str := string(b[:n])
			parts := strings.Split(str, "\n")
			if len(parts) > 1 {
				for i := 0; i < len(parts)-1; i++ {
					line += parts[i]
				}
				c <- fmt.Sprintf("%s", line)
				line = "" + parts[len(parts)-1]
			} else {
				line += parts[0]
			}
		}
	}(c)
	return c
}
