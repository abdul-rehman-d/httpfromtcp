package main

import (
	"crypto/sha256"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	proxyHandlers := map[string]server.Handler{
		"/httpbin": httpbinHandler,
	}
	server, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		for prefix, handler := range proxyHandlers {
			if strings.HasPrefix(req.RequestLine.RequestTarget, prefix) {
				handler(w, req)
				return
			}
		}

		slog.Info("got a request", "req", req)

		headers := response.GetDefaultHeaders(0)
		headers.Replace("Content-Type", "text/html")

		statusCode := response.OK
		body := []byte{}
		switch req.RequestLine.RequestTarget {
		case "/video":
			body, _ = os.ReadFile("assets/vim.mp4")
			headers.Replace("Content-Type", "video/mp4")
			headers.Replace("connection", "keep-alive")
		case "/yourproblem":
			statusCode = response.BadRequest
			body = []byte(BAD_REQUEST)
		case "/myproblem":
			statusCode = response.InternalServerError
			body = []byte(INTERNAL_SERVER_ERROR)
		default:
			body = []byte(OK)
		}

		headers.Replace("Content-Length", fmt.Sprintf("%d", len(body)))

		err := w.WriteStatusLine(statusCode)
		if err != nil {
			slog.Error("error", "error", err)
		}
		err = w.WriteHeaders(headers)
		if err != nil {
			slog.Error("error", "error", err)
		}
		_, err = w.WriteBody(body)
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

func httpbinHandler(w *response.Writer, req *request.Request) {
	h := response.GetDefaultHeaders(0)
	h.Replace("Content-Type", "text/html")
	h.Delete("Content-Length")
	h.Set("Transfer-Encoding", "chunked")
	h.Set("Trailer", "X-Content-SHA256")
	h.Set("Trailer", "X-Content-Length")

	err := w.WriteStatusLine(response.OK)
	if err != nil {
		slog.Error("error", "error", err)
	}
	err = w.WriteHeaders(h)
	if err != nil {
		slog.Error("error", "error", err)
	}

	baseUrl := "https://httpbin.org"
	endpoint := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	fullUrl := fmt.Sprintf("%s%s", baseUrl, endpoint)
	res, err := http.Get(fullUrl)

	body := []byte{}

	for {
		buff := make([]byte, 32)
		n, err := res.Body.Read(buff)
		if n == 0 || err == io.EOF {
			break
		}
		if err != nil {
			slog.Error("error", "error", err)
			break
		}

		body = append(body, buff[:n]...)

		n, err = w.WriteChunkedBody(buff[:n])
		if err != nil {
			slog.Error("error", "error", err)
			break
		}
		slog.Info("wrote bytes", "n", n)

	}
	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		slog.Error("error", "error", err)
	}

	trailers := headers.NewHeaders()

	hash := sha256.Sum256(body)
	trailers.Set("X-Content-SHA256", toStr(hash))
	trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(body)))

	err = w.WriteTrailers(*trailers)
	if err != nil {
		slog.Error("error", "error", err)
	}

}

func toStr(hash [32]byte) string {
	out := ""
	for _, b := range hash {
		out += fmt.Sprintf("%02x", b)
	}
	return out
}
