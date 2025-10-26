package main

import (
	request "github/gojogourav/http-from-scratch/Request"
	server "github/gojogourav/http-from-scratch/internals"
	"github/gojogourav/http-from-scratch/internals/response"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func main() {
	server, err := server.Serve(port, func(w io.Writer, req *request.Request) *server.HandlerError {
		if req.RequestLine.RequestTarget == "/yourproblem" {
			return &server.HandlerError{
				StatusCode: response.StatusBadRequest,
				Message:    "Your problem not my problem",
			}
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			return &server.HandlerError{
				StatusCode: response.StatusInternalServerError,
				Message:    "Woopsie my bad",
			}
		} else {
			return &server.HandlerError{
				StatusCode: response.StatusOk,
				Message:    "Everything good bro",
			}
		}

	})
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	//this enables graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
