package server

import (
	"bytes"
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
	handler  Handler
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

// Write writes a complete HTTP error response to the provided writer.
// This includes the status line, headers, and message body formatted per RFC 9112.
func (he *HandlerError) Write(w io.Writer) {
	response.WriteStatusLine(w, he.StatusCode)
	messageBytes := []byte(he.Message)
	headers := response.GetDefaultHeaders(len(messageBytes))
	response.WriteHeaders(w, headers)
	w.Write(messageBytes)
}

// Serve creates a new HTTP server listening on the specified port and starts accepting connections.
// The server runs in a separate goroutine and handles each connection concurrently.
func Serve(port int, handler Handler) (*Server, error) {
	listening, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to start server: %v", err)
	}

	server := &Server{
		listener: listening,
		handler:  handler,
	}

	go server.listen()

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

		go s.handle(conn)
	}
}

// handle processes a single HTTP connection by parsing the request and calling the provided handler.
// If the handler returns an error, it writes the error response.
// The response includes a status line with headers per RFC 9112 Section 3.
func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		handlerErr := &HandlerError{
			StatusCode: 400,
			Message:    fmt.Sprintf("Bad Request: %v", err),
		}

		handlerErr.Write(conn)
		return
	}

	buffer := bytes.NewBuffer([]byte{})
	handlerErr := s.handler(buffer, req)
	if handlerErr != nil {
		handlerErr.Write(conn)
		return
	}

	responseBody := buffer.Bytes()
	response.WriteStatusLine(conn, response.StatusOK)
	headers := response.GetDefaultHeaders(len(responseBody))
	response.WriteHeaders(conn, headers)
	conn.Write(responseBody)

	// Give the client time to read the full response before closing
	// This prevents "connection reset by peer" errors
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.CloseWrite()
	}
}
