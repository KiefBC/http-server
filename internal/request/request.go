package request

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/kiefbc/http-server-1.1/internal/headers"
)

type stateStatus int

const (
	initialized stateStatus = iota
	parsingHeaders
	parsingBody
	done
)
const bufferSize = 8

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	state       stateStatus
	Body        []byte
	bodyLength  int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

// RequestFromReader parses an HTTP request from the given io.Reader using streaming buffer management.
// It reads data in chunks, expanding the buffer as needed, and returns a parsed Request.
func RequestFromReader(r io.Reader) (*Request, error) {
	buffer := make([]byte, bufferSize)
	readToIndex := 0

	request := Request{
		Headers: headers.NewHeaders(),
		state:   initialized,
		Body:    make([]byte, 0),
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
				// Check if we're in the middle of parsing a body with Content-Length
				if request.state == parsingBody {
					if contentLengthStr, exists := request.Headers.Get("Content-Length"); exists {
						contentLength, parseErr := strconv.ParseInt(contentLengthStr, 10, 64)
						if parseErr == nil && int64(len(request.Body)) < contentLength {
							return &Request{}, fmt.Errorf("incomplete body: expected %d bytes, got %d", contentLength, len(request.Body))
						}
					}
				}
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

// parseRequestLine parses the HTTP request line from the given data bytes.
// Returns the parsed RequestLine, number of bytes consumed, and any error encountered.
func parseRequestLine(data []byte) (RequestLine, int, error) {
	// Look for CRLF line terminator (RFC 9112 Section 2.2)
	crlfIndex := strings.Index(string(data), "\r\n")
	if crlfIndex == -1 {
		// No \r\n found, need more data
		return RequestLine{}, 0, nil
	}

	// Parse request-line: method SP request-target SP HTTP-version (RFC 9112 Section 3)
	line := string(data[:crlfIndex])
	parsed := strings.Split(line, " ")
	if len(parsed) != 3 {
		return RequestLine{}, 0, fmt.Errorf("invalid request line format: expected 3 parts, got %d", len(parsed))
	}

	// Validate HTTP method format (RFC 9110 Section 9)
	method := parsed[0] // GET or POST
	if method != strings.ToUpper(method) {
		return RequestLine{}, 0, fmt.Errorf("invalid http method")
	}

	requestTarget := parsed[1] // /somewhere

	// Validate HTTP version format and support (RFC 9112 Section 3)
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

// parse processes the given data bytes, potentially in multiple steps, until the request is fully parsed or more data is needed.
// Returns the total number of bytes consumed and any error encountered.
func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != done {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return totalBytesParsed, err
		}
		if n == 0 {
			break // need more data
		}
		totalBytesParsed += n
	}
	return totalBytesParsed, nil
}

// parseSingle processes a single step of parsing based on the current state.
// It returns the number of bytes consumed in this step and any error encountered.
func (r *Request) parseSingle(data []byte) (int, error) {
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
		r.state = parsingHeaders
		return bytesRead, nil
	case parsingHeaders:
		bytesRead, headersDone, err := r.Headers.Parse(data)
		if err != nil {
			return 0, fmt.Errorf("error parsing headers: %v", err)
		}
		if bytesRead == 0 {
			return 0, nil // need more data
		}
		if headersDone {
			r.state = parsingBody
		}
		return bytesRead, nil
	case parsingBody:
		if contentLengthStr, exists := r.Headers.Get("Content-Length"); exists {
			contentLength, err := strconv.ParseInt(contentLengthStr, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid Content-Length: %v", err)
			}

			// r.Body = append(r.Body, data...)
			// r.bodyLength += len(data)

			// Only take the exact amount of bytes needed for the body
			bytesNeeded := contentLength - int64(len(r.Body))
			bytesToTake := min(int64(len(data)), bytesNeeded)
			r.Body = append(r.Body, data[:bytesToTake]...)

			// if int64(r.bodyLength) > contentLength {
			//	return 0, fmt.Errorf("Content-Length too large")
			// }

			if int64(len(r.Body)) == contentLength {
				r.state = done
			}

			// return len(data), nil
			return int(bytesToTake), nil // Return only what we consumed
		} else {
			r.state = done
			return len(data), nil
		}

	case done:
		return 0, fmt.Errorf("trying to read data in a done state")
	default:
		return 0, fmt.Errorf("invalid state")
	}
}
