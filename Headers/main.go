package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers struct {
	headers map[string]string
}

var rn = []byte("\r\n")

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

var (
	MalformedHeader = fmt.Errorf("Error Malformed Headers")
)

func IsValidToken(str string) bool {
	//     Uppercase letters: A-Z
	//     Lowercase letters: a-z
	//     Digits: 0-9
	//     Special characters: !, #, $, %, &, ', *, +, -, ., ^, _, `, |, ~

	// and at least a length of 1.
	if len(str) == 0 {
		return false
	}
	for _, ch := range str {
		if ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' || strings.ContainsRune(" !, #, $, %, &, ', *, +, -, ., ^, _, `, |, ~", ch) {
			continue
		} else {
			return false
		}
	}
	return true

}

func parseHeaders(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", MalformedHeader
	}
	key := parts[0]
	value := bytes.TrimSpace(parts[1])

	if !IsValidToken(string(key)) {
		return "", "", MalformedHeader
	}

	if bytes.HasSuffix(key, []byte(" ")) {
		return "", "", MalformedHeader
	}
	fmt.Printf("Key is - %s\nValue is %s\n", string(key), string(value))

	return string(key), string(value), nil
}

func (h *Headers) Get(key string) string {
	return h.headers[strings.ToLower(key)]
}
func (h *Headers) Set(key, value string) {
	h.headers[strings.ToLower(key)] = value
}

func (h *Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false
	for {
		idx := bytes.Index(data[read:], rn)
		if idx == -1 {
			done = false
			break
		}
		if idx == 0 {
			done = true
			read += len(rn)
			break
		}
		key, value, err := parseHeaders(data[read : idx+read])
		if err != nil {
			return 0, false, err
		}

		h.Set(key, value)
		read += idx + len(rn)

	}
	return read, done, nil
}
