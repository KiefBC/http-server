package server

import (
	"fmt"
	"github.com/kiefbc/http-server-1.1/internal/response"
	"net"
	"sync/atomic"
)

type Server struct {
	listener net.Listener
	isClosed atomic.Bool
}

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

func (s *Server) Close() error {
	closeErr := s.listener.Close()
	if closeErr != nil {
		return fmt.Errorf("failed to close server: %v", closeErr)
	}

	s.isClosed.Store(true)
	return nil
}

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

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	// Write Status Line
	err := response.WriteStatusLine(conn, response.StatusOK)
	if err != nil {
		fmt.Printf("Error writing status line: %v\n", err)
		return
	}

	// Write Headers
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
