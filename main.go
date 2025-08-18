package main

import (
	"bytes"
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

	currentLine := ""
	for {
		buffer := make([]byte, BUFFER_SIZE, BUFFER_SIZE)
		bytesRead, err := f.Read(buffer)
		if err != nil {
			break
		}

		buffer = buffer[:bytesRead]

		for {
			if idx := bytes.IndexByte(buffer, '\n'); idx >= 0 {
				currentLine += string(buffer[:idx])
				fmt.Printf("read: %s\n", currentLine)

				buffer = buffer[idx+1:]
				currentLine = ""
			} else {
				break
			}
		}

		currentLine += string(buffer)
	}
}
