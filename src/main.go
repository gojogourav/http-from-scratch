package main

import (
	"fmt"
	request "github/gojogourav/http-from-scratch/Request"
	server "github/gojogourav/http-from-scratch/internals"
	"github/gojogourav/http-from-scratch/internals/response"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func main() {
	s, err := server.Serve(port, func(w *response.Writer, req *request.Request) *server.HandlerBody {
		headers := response.GetDefaultHeaders(0)
		headers.Set("Content-Type", "text/html")
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			body := []byte(`
<html>
  <head><title>400 Bad Request</title></head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)

			headers.Set("Content-Length", fmt.Sprintf("%d", len(body)))
			w.WriteStatusLine(response.StatusBadRequest)
			w.WriteHeaders(headers)
			w.WriteBody(body)

			return &server.HandlerBody{

				StatusCode: response.StatusBadRequest,
				Message:    "Your problem not my problem",
			}
		case "/myproblem":

			body := []byte(`
<html>
  <head><title>500 Internal Server Error</title></head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
			headers.Set("Content-Length", fmt.Sprintf("%d", len(body)))
			w.WriteStatusLine(response.StatusInternalServerError)
			w.WriteHeaders(headers)
			w.WriteBody(body)

			return &server.HandlerBody{
				StatusCode: response.StatusInternalServerError,
				Message:    "Woopsie my bad",
			}
		default:
			body := []byte(`
<html>
  <head><title>200 OK</title></head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
			headers.Set("Content-Length", fmt.Sprintf("%d", len(body)))
			w.WriteStatusLine(response.StatusOk)
			w.WriteHeaders(headers)
			w.WriteBody(body)
			return &server.HandlerBody{
				StatusCode: response.StatusOk,
				Message:    "no problem bro",
			}
		}
	})

	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer s.Close()
	log.Println("Server started on port", port)

	//this enables graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
