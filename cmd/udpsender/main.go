package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

const (
	address = "localhost:42069"
)

func main() {
	serverAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving UDP address: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("starting server on... %v\n\n", serverAddr)

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error dialing UDP: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	buffer := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("-> ")
		userInput, err := buffer.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
		}
		_, err = conn.Write([]byte(userInput))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error sending data: %v\n", err)
			continue
		}
	}
}
