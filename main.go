package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const filename = "messages.txt"

func main() {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("error opening file %q: %v", filename, err)
	}
	defer file.Close()

	fmt.Printf("Reading data from %s\n", filename)
	fmt.Println("=====================================")

	ch := getLinesChannel(file)

	for elem := range ch {
		fmt.Printf("read: %s\n", elem)
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
