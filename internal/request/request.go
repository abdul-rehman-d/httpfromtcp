package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
}

var ERROR_MALFORMED_REQUEST_LINE = fmt.Errorf("malformed request line")
var ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("unsupported http version")
var ERROR_INCORRECT_METHOD = fmt.Errorf("incorrect method")
var SEPERATOR = "\r\n"
var HTTP_NAME = "HTTP"
var SLASH = '/'
var SP = " "

func (req *RequestLine) validateMethod() bool {
	return req.Method == strings.ToUpper(req.Method)
}

func (req *RequestLine) parseHttpVersion() (string, bool) {
	if req.HttpVersion == "HTTP/1.1" {
		return "1.1", true
	}
	return "", false
}

func parseRequestLine(data string) (*RequestLine, string, error) {
	idx := strings.Index(data, SEPERATOR)
	if idx == -1 {
		return nil, "", ERROR_MALFORMED_REQUEST_LINE
	}

	startLine := data[:idx]

	parts := strings.Split(startLine, SP)
	if len(parts) != 3 {
		return nil, "", ERROR_MALFORMED_REQUEST_LINE
	}
	reqLine := &RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   parts[2],
	}

	if ver, ok := reqLine.parseHttpVersion(); ok {
		reqLine.HttpVersion = ver
	} else {
		return nil, "", ERROR_UNSUPPORTED_HTTP_VERSION
	}
	if !reqLine.validateMethod() {
		return nil, "", ERROR_INCORRECT_METHOD
	}

	return reqLine, data[idx+len(SEPERATOR):], nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("unable to do io.ReadAll"),
			err,
		)
	}

	reqLine, _, err := parseRequestLine(string(data))
	if err != nil {
		return nil, err
	}

	req := &Request{
		RequestLine: *reqLine,
	}

	return req, nil
}
