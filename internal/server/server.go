package server

import (
	"fmt"
	"io"
	"net"
	"sync/atomic"

	"github.com/kiefbc/http-server-1.1/internal/request"
	"github.com/kiefbc/http-server-1.1/internal/response"
)

type Server struct {
	listener net.Listener
	isClosed atomic.Bool
}

type HandlerError struct {
	statusCode response.StatusCode
	message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

// Serve creates a new HTTP server listening on the specified port and starts accepting connections.
// The server runs in a separate goroutine and handles each connection concurrently.
func Serve(port int) (*Server, error) {
	listening, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to start server: %v", err)
	}

	server := &Server{
		listener: listening,
	}

	fmt.Printf("Server listening on %v\n", listening.Addr())

	go func() {
		defer listening.Close()
		server.listen()
	}()

	return server, nil
}

// Close gracefully shuts down the server by closing the listener and setting the closed flag.
func (s *Server) Close() error {
	closeErr := s.listener.Close()
	if closeErr != nil {
		return fmt.Errorf("failed to close server: %v", closeErr)
	}

	s.isClosed.Store(true)
	return nil
}

// listen continuously accepts new connections until the server is closed.
// Each accepted connection is handled concurrently in its own goroutine.
func (s *Server) listen() {
	for !s.isClosed.Load() {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}

		go func(c net.Conn) {
			s.handle(c)
		}(conn)
	}
}

// handle processes a single HTTP connection by sending a basic HTTP response.
// The response includes a 200 OK status line with default headers per RFC 9112 Section 3.
func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	// Write Status Line (RFC 9112 Section 3.1.2)
	err := response.WriteStatusLine(conn, response.StatusOK)
	if err != nil {
		fmt.Printf("Error writing status line: %v\n", err)
		return
	}

	// Write Headers (RFC 9112 Section 3.2)
	headers := response.GetDefaultHeaders(0)
	err = response.WriteHeaders(conn, headers)
	if err != nil {
		fmt.Printf("Error writing headers: %v\n", err)
		return
	}

	// Give the client time to read the full response before closing
	// This prevents "connection reset by peer" errors
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.CloseWrite()
	}
}
