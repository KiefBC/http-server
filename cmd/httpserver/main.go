package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kiefbc/http-server-1.1/internal/request"
	"github.com/kiefbc/http-server-1.1/internal/server"
)

const port = 42069

// handler processes HTTP requests based on the request path/target.
// Returns specific error responses for /yourproblem and /myproblem paths,
// otherwise writes a success message to the response body.
func handler(w io.Writer, req *request.Request) *server.HandlerError {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		return &server.HandlerError{
			StatusCode: 400,
			Message:    "Your problem is not my problem\n",
		}
	case "/myproblem":
		return &server.HandlerError{
			StatusCode: 500,
			Message:    "Woopsie, my bad\n",
		}
	default:
		w.Write([]byte("All good, frfr\n"))
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
