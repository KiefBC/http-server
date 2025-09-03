package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

const (
	bytesToRead = 8
	port        = ":42069"
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

		lineChannel := getLinesChannel(conn)

		for line := range lineChannel {
			fmt.Printf("%s\n", line)
		}

		fmt.Printf("\n<- client disconnected\n\n")
	}
}

func getLinesChannel(r io.ReadCloser) <-chan string {
	buffer := make([]byte, bytesToRead) // only one param defaults capacity to that param
	ch := make(chan string)

	go func() {
		defer close(ch)
		defer r.Close()

		var currentLine string

		for {
			reader, err := r.Read(buffer)
			if err != nil {
				if currentLine != "" {
					ch <- currentLine
				}
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Fprintf(os.Stderr, "Could not read file: %v\n", err)
				break
			}

			currentLine += string(buffer[:reader])
			parts := strings.Split(currentLine, "\n")

			for i := 0; i < len(parts)-1; i++ {
				ch <- parts[i]
			}

			currentLine = parts[len(parts)-1]
		}
	}()

	return ch
}
