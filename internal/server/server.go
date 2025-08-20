package server

import (
	"fmt"
	"httpfromtcp/internal/response"
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
	err := response.WriteStatusLine(conn, response.OK)
	if err != nil {
		slog.Error("error", "failed to write response status line", err)
	}
	err = response.WriteHeaders(conn, response.GetDefaultHeaders(0))
	if err != nil {
		slog.Error("error", "failed to write response headers", err)
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
