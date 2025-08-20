package request

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strings"
)

var ERROR_MALFORMED_REQUEST_LINE = fmt.Errorf("malformed request line")
var ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("unsupported http version")
var ERROR_INCORRECT_METHOD = fmt.Errorf("incorrect method")
var SEPERATOR = []byte("\r\n")
var SP = []byte(" ")

type parserState string

const (
	StateInit         parserState = "init"
	StateDone         parserState = "done"
	StateParseHeaders parserState = "headers"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (req *RequestLine) validateMethod() bool {
	return req.Method == strings.ToUpper(req.Method)
}

func (req *RequestLine) parseHttpVersion() (string, bool) {
	if req.HttpVersion == "HTTP/1.1" {
		return "1.1", true
	}
	return "", false
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	state       parserState
}

func newRequest() *Request {
	return &Request{
		state: StateInit,
	}
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		switch r.state {
		case StateInit:
			rl, n, err := parseRequestLine(data[read:])
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break outer
			}
			read += n
			r.RequestLine = *rl
			r.state = StateParseHeaders
		case StateParseHeaders:
			headers := headers.NewHeaders()
			fmt.Println(string(data[read:]))
			n, done, err := headers.Parse(data[read:])
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break outer
			}
			if !done {
				break outer
			}
			read += n
			r.Headers = headers
			r.state = StateDone
		case StateDone:
			break outer
		}
	}
	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, SEPERATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	read := idx + len(SEPERATOR)
	startLine := data[:idx]

	parts := bytes.Split(startLine, SP)
	if len(parts) != 3 {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}
	reqLine := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(parts[2]),
	}

	if ver, ok := reqLine.parseHttpVersion(); ok {
		reqLine.HttpVersion = ver
	} else {
		return nil, 0, ERROR_UNSUPPORTED_HTTP_VERSION
	}
	if !reqLine.validateMethod() {
		return nil, 0, ERROR_INCORRECT_METHOD
	}

	return reqLine, read, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := newRequest()

	buf := make([]byte, 1024)
	bufLen := 0

	for !req.done() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}

		bufLen += n

		reqN, err := req.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[reqN:bufLen])
		bufLen -= reqN
	}

	return req, nil
}
