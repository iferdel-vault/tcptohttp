package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
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

	b := make([]byte, 8, 8)

	for {
		n, err := file.Read(b)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatalf("error reading file: %v", err)
			break
		}
		str := string(b[:n])
		fmt.Printf("read: %s\n", str)
	}
}
