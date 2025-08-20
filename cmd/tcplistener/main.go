package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
)

const BUFFER_SIZE = 8

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("error", "error", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error", "error", err)
		}

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("error", "error", err)
		}

		fmt.Printf(
			"Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n",
			req.RequestLine.Method,
			req.RequestLine.RequestTarget,
			req.RequestLine.HttpVersion,
		)
		fmt.Println("Headers:")
		for header := range req.Headers.Range() {
			fmt.Printf("- %s: %s\n", header, req.Headers.Get(header))
		}
		fmt.Println("Body:")
		fmt.Println(string(req.Body))
	}
}
