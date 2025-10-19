package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *RequestLine) ValidHTTP() bool {
	return r.HttpVersion == "HTTP/1.1"
}

type Request struct {
	RequestLine RequestLine
	Headers     map[string]string
	Body        []byte
}

var ERROR_BAD_REQUEST_LINE = fmt.Errorf("malformed request line")
var ERR_BAD_HEADER = fmt.Errorf("Malformed header")
var SEPERATOR = "\r\n"

type parserState string

const (
	StateInit parserState = "init"
	StateDone parserState = "done"
)

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	line = strings.TrimRight(line, "\r\n")
	return line, nil
}

func parseHeasers(r *bufio.Reader) (map[string]string, error) {
	headers := make(map[string]string)
	for {
		line, err := readLine(r)
		if err != nil {
			return nil, err
		}

		if line == "" {
			break
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, ERR_BAD_HEADER
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		headers[key] = value
	}
	return headers, nil
}

func parseRequestLine(r *bufio.Reader) (*RequestLine, error) {
	// GET / HTTP/1.1\r\nHost: localhost\r\n\r\n
	line, err := readLine(r)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, ERROR_BAD_REQUEST_LINE
	}
	rl := RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   parts[2],
	}
	if !rl.ValidHTTP() {
		return nil, ERROR_BAD_REQUEST_LINE
	}
	return &rl, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := bufio.NewReader(reader)

	rl, err := parseRequestLine(buf)

	headers, err := parseHeasers(buf)
	if err != nil {
		return nil, err
	}

	return &Request{
		RequestLine: (*rl),
		Headers:     headers,
	}, err
}
