package response

import (
	"fmt"
	headers "github/gojogourav/http-from-scratch/Headers"
	"io"
)

type StatusCode int

type Response struct {
}

const (
	StatusOk                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	statusLine := []byte{}
	switch statusCode {
	case StatusOk:
		statusLine = []byte("HTTP/1.1 200 ok\r\n")
	case StatusBadRequest:
		statusLine = ([]byte("HTTP/1.1 400 Bad Request\r\n"))
	case StatusInternalServerError:
		statusLine = ([]byte("HTTP/1.1 200 Internal Server Error\r\n"))
	default:
		return fmt.Errorf("Unrecognized Error Code")
	}

	_, err := w.Write(statusLine)
	return err
}

func GetDefaultHeaders(contentLen int) *headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}

func WriteHeaders(w io.Writer, h *headers.Headers) error {
	b := []byte{}
	h.ForEach(func(key, value string) {
		b = fmt.Appendf(b, "%s: %s\r\n", key, value)
	})
	b = fmt.Append(b, "\r\n")
	_, err := w.Write(b)
	return err
}
