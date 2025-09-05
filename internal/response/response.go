package response

import (
	"fmt"
	"io"

	"github.com/kiefbc/http-server-1.1/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

// WriteStatusLine writes an HTTP status line to the provided writer.
// The status line format follows RFC 9112 Section 3.1.2: HTTP-version SP status-code SP [reason-phrase] CRLF
func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case 200:
		_, err := fmt.Fprintf(w, "HTTP/1.1 200 OK\r\n")
		return err
	case 400:
		_, err := fmt.Fprintf(w, "HTTP/1.1 400 Bad Request\r\n")
		return err
	case 500:
		_, err := fmt.Fprintf(w, "HTTP/1.1 500 Internal Server Error\r\n")
		return err
	default:
		_, err := fmt.Fprintf(w, "HTTP/1.1 %d\r\n", statusCode)
		return err
	}
}

// GetDefaultHeaders creates a standard set of HTTP response headers.
// Includes Content-Length, Connection, and Content-Type headers per RFC 9110 recommendations.
func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	// Content-Length header (RFC 9110 Section 8.6)
	headers.Set("content-length", fmt.Sprintf("%d", contentLen))
	// Connection header (RFC 9110 Section 7.6.1)
	headers.Set("connection", "close")
	// Content-Type header (RFC 9110 Section 8.3)
	headers.Set("content-type", "text/plain")
	return headers
}

// WriteHeaders writes HTTP header fields to the provided writer.
// Each header follows RFC 9112 Section 3.2 format: field-name ":" field-value CRLF
func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", key, value)
		if err != nil {
			return err
		}
	}
	// Empty line marks end of headers section (RFC 9112 Section 3)
	_, err := fmt.Fprintf(w, "\r\n")
	return err
}
