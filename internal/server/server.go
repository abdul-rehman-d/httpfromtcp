package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"log/slog"
	"net"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	closed   bool
	listener net.Listener
	handler  Handler
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
	defer conn.Close()

	writer := response.NewResponseWriter(conn)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		writer.WriteStatusLine(response.BadRequest)
		writer.WriteHeaders(response.GetDefaultHeaders(0))
		return
	}

	s.handler(writer, req)
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener: listener,
		closed:   false,
		handler:  handler,
	}

	go s.listen()

	return s, nil
}
