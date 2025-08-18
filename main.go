package main

import (
	"fmt"
	"log"
	"os"
)

const BUFFER_SIZE = 8

func main() {
	f, err := os.Open("./messages.txt")
	if err != nil {
		log.Fatal("no messages.txt file")
	}

	buffer := make([]byte, BUFFER_SIZE, BUFFER_SIZE)

	for {
		bytesRead, err := f.Read(buffer)
		if err != nil {
			break
		}
		fmt.Printf("read: %s\n", string(buffer[:bytesRead]))
	}
}
