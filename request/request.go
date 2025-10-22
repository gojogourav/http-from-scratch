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
	part := strings.Split(rl.HttpVersion, "/")
	if len(part) != 2 {
		return false
	}
	return (part[0] == "HTTP") && (part[1] == "1.1")

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
	Body        []byte
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
		return 0, nil, nil //we send nil as we expect there is not enough data to be parsed
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

	if !(rl.ValidHTTPVersion()) {
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

		//this part was imp to understand if you fuckup here you'll
		//fuckup everywhere else
		case StateRequestLine:
			var rl *RequestLine
			consumedInStep, rl, err = r.parseRequestLine(workingData)
			if err != nil {
				return 0, err
			}
			if consumedInStep > 0 {
				r.RequestLine = *rl
				r.state = StateHeaders
				consumed += consumedInStep
				continue
			} //if consumed is 0 then it'll break

		case StateHeaders:
			headersEnd := bytes.Index(workingData, []byte(HEADER_END))
			if headersEnd == -1 {
				return consumed, nil
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
			continue

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
				r.Body = workingData[:length]
				consumedInStep = length
				r.state = StateDone
				continue
			} else {
				return consumed, nil
			}
		}

		if consumedInStep == 0 {
			break
		}
		consumed += consumedInStep

	}
	return consumed, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := newRequest()
	buf := make([]byte, 0, 4096)
	readBuf := make([]byte, 1024)

	for {
		n, readErr := reader.Read(readBuf)
		// fmt.Errorf((readErr.Error()))
		if n > 0 {
			buf = append(buf, readBuf[:n]...)
		}
		// if err != nil {
		// 	return nil, err
		// }
		consumed, parseErr := req.parse(buf)
		if parseErr != nil {
			return nil, parseErr
		}
		if consumed > 0 {
			buf = buf[consumed:]
		}
		if req.state == StateDone {
			fmt.Printf("Request parsing donee")
			break
		}

		if readErr != nil {
			if readErr == io.EOF {
				if req.state != StateDone {
					return nil, io.ErrUnexpectedEOF
				}
				break
			}
			return nil, readErr
		}

		//if readErr and consumed is nil,0 -> it'll repeat again //VERY VERY IMPORTANT TO GRASP THIS
		//THIS IS SOUL OF OUR PROGRAM
	}
	return req, nil
}
