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
		if (ch >= 'A' && ch <= 'Z') ||
			(ch >= 'a' && ch <= 'z') ||
			(ch >= '0' && ch <= '9') ||
			strings.ContainsRune("!#$%&'*+-.^_`|~", ch) {
			continue
		}
		return false
	}
	return true
}

func parseHeaders(fieldLine []byte) (string, string, error) {

	isValid := strings.Contains(string(fieldLine), ":")
	if !isValid {
		// fmt.Println("The ehader isn't valid ")
		// fmt.Println("This is fieldline - ", string(fieldLine))
		return "", "", MalformedHeader
	}
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(parts) != 2 {
		// fmt.Println("The ehader doesn't have two parts")
		return "", "", MalformedHeader
	}
	key := parts[0]
	value := bytes.TrimSpace(parts[1])
	// fmt.Printf("THIS IS AFTER TRIMMING - %s\n", value)
	if !IsValidToken(string(key)) {
		// fmt.Println("The tokens aren't valid")
		return "", "", MalformedHeader
	}

	// if bytes.HasSuffix(key, []byte(" ")) {
	// 	fmt.Println("The key doesn't have the suffix space")

	// 	return "", "", MalformedHeader
	// }
	// fmt.Printf("Key is - %s\nValue is %s\n", string(key), string(value))

	return string(key), string(value), nil
}

func (h *Headers) Get(key string) string {
	val := h.headers[strings.ToLower(key)]
	if len(bytes.TrimSpace([]byte(val))) == 0 {
		return ""
	}
	return h.headers[strings.ToLower(key)]
}
func (h *Headers) Set(key, value string) {
	name := strings.ToLower(key)
	// fmt.Printf("THE KEY ENCOUNTERED IS - %s\n", name)
	if v, ok := h.headers[name]; ok {
		h.headers[name] = fmt.Sprintf("%s, %s", v, value)
	} else {
		h.headers[name] = value
	}
	// fmt.Printf("DEBUG: Headers map now contains %d entries:\n", len(h.headers))
	// for k, v := range h.headers {
	// 	fmt.Printf("  %s: %s\n", k, v)
	// }

}
func (h *Headers) Display() string {
	for k, v := range h.headers {
		return fmt.Sprintf("%s: %s\n", k, v)
	}
	return ""
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
		lineEnd := read + idx
		line := data[read:lineEnd]
		if len(line) == 0 {
			done = true
			read += len(rn)
			break
		}
		key, value, err := parseHeaders(line)
		key = string(bytes.TrimSpace([]byte(key)))
		if err != nil {
			// println("error idhar h kya")
			return 0, false, err
		}

		h.Set(key, value)
		read = lineEnd + len(rn)

	}
	return read, done, nil
}

func (h *Headers) ForEach(f func(key, value string)) {
	for k, v := range h.headers {
		f(k, v)
	}
}
