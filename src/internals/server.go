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
type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func runConnection(s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()
	if s.Closed {
		return
	}
	headers := response.GetDefaultHeaders(0)

	r, err := request.RequestFromReader(conn)
	if err != nil {
		response.WriteStatusLine(conn, response.StatusBadRequest)
		response.WriteHeaders(conn, headers)
		return
	}

	writer := bytes.NewBuffer([]byte{})
	handlerError := s.Handler(writer, r)

	if handlerError != nil {

		response.WriteStatusLine(conn, handlerError.StatusCode)
		response.WriteHeaders(conn, headers)
		conn.Write([]byte(handlerError.Message))
		return
	}

	body := writer.Bytes()
	headers.Set("Content-Length", fmt.Sprintf("%d", body))

	response.WriteStatusLine(conn, response.StatusOk)
	response.WriteHeaders(conn, headers)
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
