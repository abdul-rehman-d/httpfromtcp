package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

type StatusCode string

const (
	OK                  StatusCode = "200 OK"
	BadRequest          StatusCode = "400 Bad Request"
	InternalServerError StatusCode = "500 Internal Server Error"
)

type Response struct{}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	str := fmt.Sprintf("HTTP/1.1 %s\r\n", statusCode)
	return writeAndHandleErrs(w, str)
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k := range headers.Range() {
		v := headers.Get(k)

		str := fmt.Sprintf("%s: %s\r\n", k, v)
		if err := writeAndHandleErrs(w, str); err != nil {
			return err
		}
	}
	// END OF HEADERS: EMPTY HEADER
	if err := writeAndHandleErrs(w, "\r\n"); err != nil {
		return err
	}

	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()

	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return *h
}

func writeAndHandleErrs(w io.Writer, s string) error {
	n, err := w.Write([]byte(s))

	if n != len(s) {
		return fmt.Errorf("could not write full response line")
	}
	if err != nil {
		return err
	}

	return nil
}
