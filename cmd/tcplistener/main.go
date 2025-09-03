package main

import (
	"fmt"
	"github.com/kiefbc/http-server-1.1/internal/request"
	"net"
	"os"
)

const (
	port = ":42069"
)

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not start server: %v\n", err)
		return
	}
	defer listener.Close()

	fmt.Printf("starting server on... %v\n\n", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error accepting connection: %v\n", err)
			continue
		}

		fmt.Printf("-> addr %v has connected\n\n", conn.RemoteAddr())

		request, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading request: %v\n", err)
			conn.Close()
			continue
		}

		fmt.Printf("-> received request: %+v\n\n", request)

		fmt.Println("Request line:")
		fmt.Printf("- Method: %v\n", request.RequestLine.Method)
		fmt.Printf("- Target: %v\n", request.RequestLine.RequestTarget)
		fmt.Printf("- Version: %v\n", request.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for k, v := range request.Headers {
			fmt.Printf("- %v: %v\n", k, v)
		}

		conn.Close()
		fmt.Printf("\n<- client disconnected\n\n")
	}
}
