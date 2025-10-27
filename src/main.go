package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	request "github/gojogourav/http-from-scratch/Request"
	server "github/gojogourav/http-from-scratch/internals"
	"github/gojogourav/http-from-scratch/internals/response"
)

const port = 42069

func main() {
	s, err := server.Serve(port, func(w *response.Writer, req *request.Request) *server.HandlerBody {
		headers := response.GetDefaultHeaders(0)
		headers.Set("Content-Type", "text/html")

		path := req.RequestLine.RequestTarget

		// ✅ Use if / else if instead of switch
		switch path {
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

		case "/chunked":
			log.Println("Proxying httpbin stream...")

			if _, err := fmt.Fprint(w, "HTTP/1.1 200 OK\r\n"); err != nil {
				log.Println("Error writing status line:", err)
				return &server.HandlerBody{StatusCode: response.StatusInternalServerError}
			}

			if err := response.ProxyHTTPinStream(w, 10); err != nil {
				log.Println("Error proxying httpbin stream:", err)
				return &server.HandlerBody{
					StatusCode: response.StatusInternalServerError,
					Message:    "Failed to proxy httpbin stream",
				}
			}

			return &server.HandlerBody{
				StatusCode: response.StatusOk,
				Message:    "Chunked httpbin stream sent",
			}
		case "/video":
			f, err := os.ReadFile("assets/vim.mp4")
			if err != nil {
				log.Println("Error reading asset file:", err)
				return &server.HandlerBody{
					StatusCode: response.StatusInternalServerError,
					Message:    "Failed to read asset",
				}
			}

			headers.Delete("Content-Type") // remove text/plain
			headers.Set("Content-Type", "video/mp4")
			headers.Set("Content-Length", fmt.Sprintf("%d", len(f)))
			headers.Set("Connection", "close")

			w.WriteStatusLine(response.StatusOk)
			w.WriteHeaders(headers)

			w.WriteBody(f)

			return &server.HandlerBody{
				StatusCode: response.StatusOk,
				Message:    "Video served successfully",
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

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
