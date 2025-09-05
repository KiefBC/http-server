package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kiefbc/http-server-1.1/internal/request"
	"github.com/kiefbc/http-server-1.1/internal/response"
	"github.com/kiefbc/http-server-1.1/internal/server"
)

const port = 42069

// handler demonstrates the new response.Writer capabilities:
// 1. Writing raw []byte HTML content for all responses
// 2. Setting custom Content-Type headers to text/html
// 3. Proper order validation with WriteStatusLine -> WriteHeaders -> WriteBody
func handler(w *response.Writer, req *request.Request) *server.HandlerError {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		htmlContent := []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)

		responseHeaders := response.GetDefaultHeaders(len(htmlContent))
		responseHeaders.Replace("content-type", "text/html")
		responseHeaders.Replace("content-length", fmt.Sprintf("%d", len(htmlContent)))

		w.WriteStatusLine(response.StatusBadRequest)
		w.WriteHeaders(responseHeaders)
		w.WriteBody(htmlContent)

		return nil

	case "/myproblem":
		htmlContent := []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)

		responseHeaders := response.GetDefaultHeaders(len(htmlContent))
		responseHeaders.Replace("content-type", "text/html")
		responseHeaders.Replace("content-length", fmt.Sprintf("%d", len(htmlContent)))

		w.WriteStatusLine(response.StatusInternalServerError)
		w.WriteHeaders(responseHeaders)
		w.WriteBody(htmlContent)

		return nil

	default:
		htmlContent := []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)

		responseHeaders := response.GetDefaultHeaders(len(htmlContent))
		responseHeaders.Replace("content-type", "text/html")
		responseHeaders.Replace("content-length", fmt.Sprintf("%d", len(htmlContent)))

		w.WriteStatusLine(response.StatusOK)
		w.WriteHeaders(responseHeaders)
		w.WriteBody(htmlContent)

		return nil
	}
}

// main starts an HTTP server that listens on port 42069 and handles graceful shutdown.
func main() {
	srv, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
