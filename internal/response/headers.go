package response

import (
	"fmt"
	"io"

	"github.com/kiefbc/http-server-1.1/internal/headers"
)

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

// GetChunkedHeaders creates headers for chunked transfer encoding responses.
// Sets Transfer-Encoding: chunked and omits Content-Length per RFC 9112 Section 7.1.
// Content-Length and Transfer-Encoding are mutually exclusive.
func GetChunkedHeaders() headers.Headers {
	headers := headers.NewHeaders()
	// Transfer-Encoding header for chunked responses (RFC 9112 Section 7.1)
	headers.Set("transfer-encoding", "chunked")
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
