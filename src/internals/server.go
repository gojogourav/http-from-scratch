package server

import (
	"bytes"
	"fmt"
	request "github/gojogourav/http-from-scratch/Request"
	"github/gojogourav/http-from-scratch/internals/response"
	"io"
	"net"
)

type Server struct {
	Closed  bool
	Handler Handler
}
type HandlerBody struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w *response.Writer, req *request.Request) *HandlerBody

func runConnection(s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()

	w := &response.Writer{
		Writer: conn,
	}

	if s.Closed {
		return
	}
	headers := response.GetDefaultHeaders(0)

	r, err := request.RequestFromReader(conn)
	if err != nil {
		w.WriteStatusLine(response.StatusBadRequest)
		w.WriteHeaders(headers)
		return
	}

	writer := bytes.NewBuffer([]byte{})
	handlerBody := s.Handler(w, r)

	var body []byte = nil
	var status response.StatusCode = response.StatusOk
	if handlerBody != nil {
		status = handlerBody.StatusCode
		body = []byte(handlerBody.Message)

	} else {
		status = response.StatusOk
		body = writer.Bytes()
	}

	headers.Set("Content-Length", fmt.Sprintf("%d", len(body)))
	w.WriteStatusLine(status)
	w.WriteHeaders(headers)
	conn.Write(body)
}

func runServer(s *Server, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if s.Closed {
			return
		}
		if err != nil {
			return
		}
		go runConnection(s, conn)
	}
}

func Serve(port uint16, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	server := &Server{
		Closed:  false,
		Handler: handler,
	}
	go runServer(server, listener)
	return server, nil
}

func (s *Server) Close() error {
	s.Closed = true
	return nil
}
