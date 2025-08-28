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

type WriterState = string

const (
	InitState     WriterState = "init"
	HeaderState   WriterState = "headers"
	BodyState     WriterState = "body"
	TrailersState WriterState = "trailers"
	DoneState     WriterState = "done"
)

type Writer struct {
	writer      io.Writer
	writerState WriterState
}

func NewResponseWriter(conn io.Writer) *Writer {
	return &Writer{
		writer:      conn,
		writerState: InitState,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != InitState {
		return fmt.Errorf("not in correct state")
	}
	str := fmt.Sprintf("HTTP/1.1 %s\r\n", statusCode)
	return w.writeStringAndHandleErrs(str, HeaderState)
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState != HeaderState {
		return fmt.Errorf("not in correct state")
	}
	for k := range headers.Range() {
		v := headers.Get(k)

		str := fmt.Sprintf("%s: %s\r\n", k, v)
		if err := w.writeStringAndHandleErrs(str, HeaderState); err != nil {
			return err
		}
	}
	// END OF HEADERS: EMPTY HEADER
	if err := w.writeStringAndHandleErrs("\r\n", BodyState); err != nil {
		return err
	}

	return nil
}

func (w *Writer) WriteTrailers(headers headers.Headers) error {
	if w.writerState != TrailersState {
		return fmt.Errorf("not in correct state")
	}
	for k := range headers.Range() {
		v := headers.Get(k)

		str := fmt.Sprintf("%s: %s\r\n", k, v)
		if err := w.writeStringAndHandleErrs(str, HeaderState); err != nil {
			return err
		}
	}
	// END OF TRAILERS: EMPTY HEADER
	if err := w.writeStringAndHandleErrs("\r\n", DoneState); err != nil {
		return err
	}

	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != BodyState {
		return 0, fmt.Errorf("not in correct state")
	}
	n, err := w.writer.Write(p)
	if err == nil && n == len(p) {
		w.writerState = DoneState
	}
	return n, err
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.writerState != BodyState {
		return 0, fmt.Errorf("not in correct state")
	}
	var toWrite []byte
	toWrite = fmt.Appendf(toWrite, "%x\r\n", len(p))
	toWrite = append(toWrite, p...)
	toWrite = fmt.Append(toWrite, "\r\n")
	n, err := w.writer.Write(toWrite)
	return n, err
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.writerState != BodyState {
		return 0, fmt.Errorf("not in correct state")
	}
	n, err := w.writer.Write([]byte("0\r\n"))
	if err == nil && n == 3 {
		w.writerState = TrailersState
	}
	return n, err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()

	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return *h
}

func (w *Writer) writeStringAndHandleErrs(s string, nextState WriterState) error {
	n, err := w.writer.Write([]byte(s))

	if n != len(s) {
		return fmt.Errorf("could not write full response line")
	}
	if err != nil {
		return err
	}

	w.writerState = nextState

	return nil
}
