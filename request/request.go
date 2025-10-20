package main

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

func (rl *RequestLine) ValidHTTPVersion() bool {
	return rl.HttpVersion == "HTTP/1.1"
}

type parserState int

const (
	StateRequestLine parserState = iota //0
	StateHeaders                        //1
	StateBody                           //2
	StateDone                           //3
)

type Request struct {
	RequestLine RequestLine
	Headers     map[string]string
	body        []byte
	state       parserState
}

var (
	ErrMalformedRequestLine = fmt.Errorf("Malformed request line")
	ErrInvalidContentLength = fmt.Errorf("Invalid Content-Length value")
	ErrMalformedHeader      = fmt.Errorf("Malformed header")
)

const (
	SEPERATOR  = "\r\n"
	HEADER_END = SEPERATOR + SEPERATOR
)

func newRequest() *Request {
	return &Request{
		Headers: make(map[string]string),
		state:   StateRequestLine,
	}
}

func (r *Request) parseRequestLine(data []byte) (int, *RequestLine, error) {
	lineEnd := bytes.Index(data, []byte(SEPERATOR))
	if lineEnd == -1 {
		return 0, nil, nil
	}

	line := string(data[:lineEnd])
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return 0, nil, ErrMalformedRequestLine
	}

	rl := &RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   parts[2],
	}

	if !rl.ValidHTTPVersion() {
		return 0, nil, fmt.Errorf("unsupported HTTP version: %s", rl.HttpVersion)
	}

	return lineEnd + len(SEPERATOR), rl, nil
}

func (r *Request) parse(data []byte) (int, error) {
	consumed := 0
	for r.state != StateDone {
		if consumed >= len(data) {
			break
		}

		workingData := data[consumed:]
		consumedInStep := 0
		var err error

		switch r.state {
		case StateRequestLine:
			var rl *RequestLine
			consumedInStep, rl, err = r.parseRequestLine(workingData)
			if err != nil {
				return 0, err
			}
			if consumedInStep > 0 {
				r.RequestLine = *rl
				r.state = StateHeaders
			}

		case StateHeaders:
			headersEnd := bytes.Index(workingData, []byte(HEADER_END))
			if headersEnd == -1 {
				break
			}

			headerBlock := string(workingData[:headersEnd])
			lines := strings.Split(headerBlock, SEPERATOR)

			for _, line := range lines {
				if line == "" {
					continue
				}
				parts := strings.SplitN(line, ":", 2)
				if len(parts) != 2 {
					return 0, ErrMalformedHeader
				}

				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				r.Headers[key] = value
			}
			consumedInStep = headersEnd + len(HEADER_END)
			r.state = StateBody

		case StateBody:
			contentLengthVal, ok := r.Headers["Content-Length"]
			if !ok || contentLengthVal == "0" {
				r.state = StateDone
				continue
			}

			length, err := strconv.Atoi(contentLengthVal)
			if err != nil {
				return 0, fmt.Errorf("%w : %s", ErrInvalidContentLength, err.Error())
			}

			if length <= len(workingData) {
				r.body = workingData[:length]
				consumedInStep = length
				r.state = StateDone
			} else {
				break
			}
		}

		if consumedInStep == 0 {
			break
		} else {
			consumed += consumedInStep
		}
	}
	return consumed, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := newRequest()
	buf := make([]byte, 0, 4096)
	readBuf := make([]byte, 1024)

	for req.state != StateDone {
		consumed, err := req.parse(buf)
		if err != nil {
			return nil, err
		}
		if consumed > 0 {
			buf = buf[consumed:]
		}

		if req.state == StateDone {
			break
		}

		n, err := reader.Read(readBuf)
		if n > 0 {
			buf = append(buf, readBuf[:n]...)
		}
		if err == io.EOF {
			if req.state != StateDone {
				if _, pErr := req.parse(buf); pErr != nil {
					return nil, pErr
				}
				if req.state != StateDone {
					return nil, io.ErrUnexpectedEOF
				}
			}
			break
		}
		return nil, err
	}
	return req, nil

}
