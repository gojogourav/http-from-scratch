package response

import (
	"fmt"
	headers "github/gojogourav/http-from-scratch/Headers"
	"io"
	"net/http"
)

type StatusCode int

type Response struct {
	StatusCode StatusCode
	Headers    headers.Headers
	body       []byte
}

type Writer struct {
	io.Writer
	Headers *headers.Headers
}

func ProxyHTTPinStream(w io.Writer, count int) error {
	url := fmt.Sprintf("https://httpbin.org/stream/%d", count)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch httpbin: %w", err)
	}
	defer resp.Body.Close()

	_, err = fmt.Fprint(w,
		"HTTP/1.1 200 OK\r\n"+
			"Content-Type: application/json\r\n"+
			"Transfer-Encoding: chunked\r\n"+
			"Connection: close\r\n\r\n")
	if err != nil {
		return err
	}

	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			chunk := fmt.Sprintf("%x\r\n%s\r\n", n, buf[:n])
			if _, werr := w.Write([]byte(chunk)); werr != nil {
				return werr
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read error: %w", err)
		}
	}

	_, err = fmt.Fprint(w, "0\r\n\r\n")
	return err
}

const (
	StatusOk                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	statusLine := []byte{}
	switch statusCode {
	case StatusOk:
		statusLine = []byte("HTTP/1.1 200 ok\r\n")
	case StatusBadRequest:
		statusLine = ([]byte("HTTP/1.1 400 Bad Request\r\n"))
	case StatusInternalServerError:
		statusLine = ([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
	default:
		return fmt.Errorf("Unrecognized Error Code")
	}

	_, err := w.Writer.Write(statusLine)
	return err
}
func (w *Writer) WriteHeaders(h *headers.Headers) error {
	b := []byte{}
	h.ForEach(func(key, value string) {
		b = fmt.Appendf(b, "%s: %s\r\n", key, value)
	})
	b = fmt.Append(b, "\r\n")
	_, err := w.Writer.Write(b)
	return err
}
func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func GetDefaultHeaders(contentLen int) *headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}
