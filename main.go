package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	textFile    = "messages.txt"
	bytesToRead = 8
)

func main() {
	file, err := os.Open(textFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open file: %v\n", err)
		return
	}

	lineChannel := getLinesChannel(file)

	for line := range lineChannel {
		fmt.Printf("read: %s\n", line)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	buffer := make([]byte, bytesToRead) // only one param defaults capacity to that param
	ch := make(chan string)

	go func() {
		defer close(ch)
		defer f.Close()

		var currentLine string

		for {
			reader, err := f.Read(buffer)
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
