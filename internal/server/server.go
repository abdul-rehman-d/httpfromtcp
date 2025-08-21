package server

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"log/slog"
	"net"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

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
	req, err := request.RequestFromReader(conn)
	if err != nil {
		// TOOD
		conn.Close()
		return
	}

	statusCode := response.OK
	writer := bytes.NewBuffer([]byte{})

	handlerErr := s.handler(writer, req)

	if handlerErr != nil {
		writer.Reset()
		writer.WriteString(handlerErr.Message)
		statusCode = handlerErr.StatusCode
	}

	err = response.WriteStatusLine(conn, statusCode)

	if err != nil {
		slog.Error("error", "failed to write response status line", err)
		conn.Close()
		return
	}

	err = response.WriteHeaders(conn, response.GetDefaultHeaders(writer.Len()))

	if err != nil {
		slog.Error("error", "failed to write response headers", err)
	}

	conn.Write(writer.Bytes())

	conn.Close()
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
