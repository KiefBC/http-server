package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/kiefbc/http-server-1.1/internal/headers"
	"github.com/kiefbc/http-server-1.1/internal/request"
	"github.com/kiefbc/http-server-1.1/internal/response"
	"github.com/kiefbc/http-server-1.1/internal/server"
)

const port = 42069

func handler(w *response.Writer, req *request.Request) *server.HandlerError {
	switch req.RequestLine.RequestTarget {
	case "/video":
		videoFile, err := os.ReadFile("assets/vim.mp4")
		if err != nil {
			return &server.HandlerError{
				StatusCode: 500,
				Message:    fmt.Sprintf("Failed to read video file: %v", err),
			}
		}

		responseHeaders := response.GetDefaultHeaders(len(videoFile))
		responseHeaders.Replace("content-type", "video/mp4")
		responseHeaders.Replace("content-length", fmt.Sprintf("%d", len(videoFile)))

		w.WriteStatusLine(response.StatusOK)
		w.WriteHeaders(responseHeaders)
		w.WriteBody(videoFile)

		return nil
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

	case "/chunked":
		chunkedHeaders := response.GetChunkedHeaders()
		chunkedHeaders.Replace("content-type", "text/html")

		w.WriteStatusLine(response.StatusOK)
		w.WriteHeaders(chunkedHeaders)

		chunk1 := []byte(`<html>
  <head>
    <title>Chunked Response</title>
  </head>
  <body>`)

		chunk2 := []byte(`    <h1>Get Chunk'd!</h1>
    <p>This content is being sent in chunks!</p>`)

		chunk3 := []byte(`    <p>Each chunk gets a hex size prefix.</p>
    <p>Perfect for streaming data!</p>
  </body>
</html>`)

		w.WriteChunkedBody(chunk1)
		w.WriteChunkedBody(chunk2)
		w.WriteChunkedBody(chunk3)

		w.WriteChunkedBodyDone()

		return nil

	default:
		// Check if this is a /httpbin proxy request
		if after, ok := strings.CutPrefix(req.RequestLine.RequestTarget, "/httpbin/"); ok {
			// Extract the path after /httpbin/ to proxy to httpbin.org
			proxyPath := after
			proxyURL := fmt.Sprintf("https://httpbin.org/%s", proxyPath)

			fmt.Printf("Proxying request to: %s\n", proxyURL)

			resp, err := http.Get(proxyURL)
			if err != nil {
				return &server.HandlerError{
					StatusCode: 500,
					Message:    fmt.Sprintf("Proxy request failed: %v", err),
				}
			}
			defer resp.Body.Close()

			chunkedHeaders := response.GetChunkedHeaders()
			// Copy content-type from upstream response if present
			if contentType := resp.Header.Get("Content-Type"); contentType != "" {
				chunkedHeaders.Replace("content-type", contentType)
			}

			chunkedHeaders.Set("trailer", "X-Content-SHA256, X-Content-Length")

			w.WriteStatusLine(response.StatusCode(resp.StatusCode))
			w.WriteHeaders(chunkedHeaders)

			var fullBody []byte
			buffer := make([]byte, 8)
			for {
				n, err := resp.Body.Read(buffer)
				if n > 0 {
					fmt.Printf("Read %d bytes, writing chunk\n", n)
					fullBody = append(fullBody, buffer[:n]...)
					w.WriteChunkedBody(buffer[:n])
				}
				if err != nil {
					break
				}
			}

			hash := sha256.Sum256(fullBody)

			w.WriteChunkedBodyDone()

			trailerHeaders := make(headers.Headers)
			trailerHeaders.Replace("x-content-sha256", fmt.Sprintf("%x", hash))
			trailerHeaders.Replace("x-content-length", fmt.Sprintf("%d", len(fullBody)))
			w.WriteTrailers(trailerHeaders)
			w.WriteTrailersDone()

			fmt.Println("Proxy streaming completed")

			return nil
		}

		// Default non-proxy response
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
