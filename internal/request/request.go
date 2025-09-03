package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type stateStatus int

const (
	initialized stateStatus = iota
	done
)
const bufferSize = 8

type Request struct {
	RequestLine RequestLine
	state       stateStatus
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(r io.Reader) (*Request, error) {
	buffer := make([]byte, bufferSize)
	readToIndex := 0

	request := Request{
		state: initialized,
	}

	for request.state != done {
		if readToIndex >= len(buffer) {
			// Expand the buffer if needed
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
		}

		bytesRead, err := r.Read(buffer[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				request.state = done
				break
			}
			return &Request{}, fmt.Errorf("error reading data: %w", err)
		}

		readToIndex += bytesRead

		parsedBytes, err := request.parse(buffer[:readToIndex])
		if err != nil {
			return &Request{}, fmt.Errorf("error parsing request: %v", err)
		}

		copy(buffer, buffer[parsedBytes:])
		readToIndex -= parsedBytes
	}

	return &request, nil
}

func parseRequestLine(data []byte) (RequestLine, int, error) {
	crlfIndex := strings.Index(string(data), "\r\n")
	if crlfIndex == -1 {
		// No \r\n found, need more data
		return RequestLine{}, 0, nil
	}

	line := string(data[:crlfIndex])
	parsed := strings.Split(line, " ")
	if len(parsed) != 3 {
		return RequestLine{}, 0, fmt.Errorf("invalid request line format: expected 3 parts, got %d", len(parsed))
	}

	method := parsed[0] // GET or POST
	if method != strings.ToUpper(method) {
		return RequestLine{}, 0, fmt.Errorf("invalid http method")
	}

	requestTarget := parsed[1] // /somewhere

	httpVersion := parsed[2] // HTTP/1.1
	httpVersionParts := strings.Split(httpVersion, "/")
	if len(httpVersionParts) != 2 {
		return RequestLine{}, 0, fmt.Errorf("invalid HTTP version format: %v", httpVersion)
	}

	version := strings.Split(httpVersion, "/")[1]
	if version != "1.1" {
		return RequestLine{}, 0, fmt.Errorf("only HTTP/1.1 is supported, got HTTP/%v", version)
	}

	requestLine := RequestLine{
		HttpVersion:   version,
		RequestTarget: requestTarget,
		Method:        method,
	}

	// Return the number of bytes consumed (including \r\n)
	return requestLine, crlfIndex + 2, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case initialized:
		requestLine, bytesRead, err := parseRequestLine(data)
		if err != nil {
			return 0, fmt.Errorf("error parsing request line: %v", err)
		}
		if bytesRead == 0 {
			return 0, nil // need more data
		}

		r.RequestLine = requestLine
		r.state = done
		return bytesRead, nil
	case done:
		return 0, fmt.Errorf("trying to read data in a done state")
	default:
		return 0, fmt.Errorf("invalid state")
	}
}
