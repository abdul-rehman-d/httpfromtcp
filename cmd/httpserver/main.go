package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

const (
	BAD_REQUEST = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>
`
	INTERNAL_SERVER_ERROR = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>
`
	OK = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>
`
)

func main() {
	server, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		statusCode := response.OK
		body := ""
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			statusCode = response.BadRequest
			body = BAD_REQUEST
		case "/myproblem":
			statusCode = response.InternalServerError
			body = INTERNAL_SERVER_ERROR
		default:
			body = OK
		}

		headers := response.GetDefaultHeaders(0)
		headers.Replace("Content-Type", "text/html")
		headers.Replace("Content-Length", fmt.Sprintf("%d", len(body)))

		err := w.WriteStatusLine(statusCode)
		if err != nil {
			slog.Error("error", "error", err)
		}
		err = w.WriteHeaders(headers)
		if err != nil {
			slog.Error("error", "error", err)
		}
		_, err = w.WriteBody([]byte(body))
		if err != nil {
			slog.Error("error", "error", err)
		}
	})
	if err != nil {
		slog.Error("could not start the server", "error", err)
		return
	}
	defer server.Close()
	slog.Info("server started", "port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	slog.Info("\nserver stopped gracefully")
}
