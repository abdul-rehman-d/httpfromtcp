package headers

import (
	"bytes"
	"fmt"
	"strings"
)

var ERROR_MALFORMED_HEADER = fmt.Errorf("malformed header")
var SEPERATOR = []byte("\r\n")
var COLON = byte(':')
var SP = " "

type Headers struct {
	headers map[string]string
}

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

func (h *Headers) Get(key string) string {
	return h.headers[strings.ToLower(key)]
}

func (h *Headers) Set(key string, value string) error {
	key = strings.ToLower(key)
	if !isValidToken([]byte(key)) {
		return fmt.Errorf("invalid header %s", key)

	}
	h.headers[strings.ToLower(key)] = value
	return nil
}

func (h *Headers) Parse(data []byte) (n int, done bool, err error) {
	n = 0
	done = false

	for {

		idx := bytes.Index(data, SEPERATOR)

		if idx == -1 {
			break
		}

		if idx == 0 {
			done = true
			break
		}

		read := (idx + len(SEPERATOR))
		n += read

		key, value, ok := parseHeader(data[:idx])
		if !ok {
			return 0, false, ERROR_MALFORMED_HEADER
		}

		data = data[read:]

		err = h.Set(key, value)
		if err != nil {
			return 0, false, err
		}
	}

	return n, done, nil
}

func parseHeader(data []byte) (key string, value string, ok bool) {
	colonIdx := bytes.IndexByte(data, COLON)
	if colonIdx == -1 {
		return "", "", false
	}

	key = string(data[:colonIdx])
	if len(key) == 0 || strings.HasSuffix(key, SP) {
		return "", "", false
	}
	key = strings.TrimLeft(key, SP)

	value = string(data[colonIdx+1:])
	value = strings.Trim(value, SP)

	return key, value, true
}

func isValidToken(str []byte) bool {
	for _, ch := range str {
		found := false
		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			found = true
		} else {
			switch ch {
			case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
				found = true
			}
		}

		if !found {
			return false
		}
	}
	return true
}
