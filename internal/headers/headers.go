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

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
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

		h[key] = value
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
