package headers

import (
	"bytes"
	"fmt"
	"iter"
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

func (h *Headers) Set(key string, value string) {
	key = strings.ToLower(key)
	if v, ok := h.headers[key]; ok {
		value = fmt.Sprintf("%s, %s", v, value)
	}
	h.headers[key] = value
}

func (h *Headers) Replace(key string, value string) {
	key = strings.ToLower(key)
	h.headers[key] = value
}

func (h *Headers) Range() iter.Seq[string] {
	return func(yield func(string) bool) {
		for k := range h.headers {
			if !yield(k) {
				return
			}
		}
	}
}

func (h *Headers) Parse(data []byte) (n int, done bool, err error) {
	n = 0
	done = false

	for {

		idx := bytes.Index(data, SEPERATOR)

		if idx == -1 {
			break
		}

		read := (idx + len(SEPERATOR))
		n += read

		if idx == 0 {
			done = true
			break
		}

		key, value, ok := parseHeader(data[:idx])
		if !ok {
			return 0, false, ERROR_MALFORMED_HEADER
		}

		data = data[read:]

		if !isValidToken([]byte(key)) {
			return 0, false, fmt.Errorf("invalid header %s", key)

		}

		h.Set(key, value)
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
