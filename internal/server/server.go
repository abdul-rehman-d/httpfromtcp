package server

import (
	"fmt"
	"log/slog"
	"net"
)

type Server struct {
	closed   bool
	listener net.Listener
}

func (s *Server) Close() error {
	s.closed = true
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if s.closed {
			return
		}

		if err != nil {
			slog.Error("error", "failed to accept connection", err)
			continue
		}

		go s.handle(conn)

	}

}

func (s *Server) handle(conn net.Conn) {
	dataToWrite := []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello World!")

	n, err := conn.Write(dataToWrite)
	if n != len(dataToWrite) {
		slog.Error("error failed to write data")
	}
	if err != nil {
		slog.Error("error", "failed to write data", err)
	}

	conn.Close()
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener: listener,
		closed:   false,
	}

	go s.listen()

	return s, nil
}
