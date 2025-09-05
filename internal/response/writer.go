package response

import (
	"fmt"
	"io"

	"github.com/kiefbc/http-server-1.1/internal/headers"
)

// writerState tracks the current state of response writing to ensure proper order
type writerState int

const (
	stateInit writerState = iota
	stateStatusWritten
	stateHeadersWritten
	stateBodyWritten
	stateChunkedWriting
	stateChunkedDone
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusNotFound            StatusCode = 404
	StatusForbidden           StatusCode = 403
	StatusMethodNotAllowed    StatusCode = 405
	StatusInternalServerError StatusCode = 500
)

// Writer encapsulates HTTP response writing functionality.
// Provides control over status line, headers, and body content with state validation.
type Writer struct {
	writer io.Writer
	state  writerState
}

// NewWriter creates a new response Writer that writes to the provided io.Writer.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		state:  stateInit,
	}
}

// WriteStatusLine writes an HTTP status line using the Writer's internal writer.
// Must be called first before WriteHeaders or WriteBody.
// The status line format follows RFC 9112 Section 3.1.2: HTTP-version SP status-code SP [reason-phrase] CRLF
func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != stateInit {
		return fmt.Errorf("WriteStatusLine called out of order - must be called first")
	}

	var err error
	switch statusCode {
	case 200:
		_, err = fmt.Fprintf(w.writer, "HTTP/1.1 200 OK\r\n")
	case 400:
		_, err = fmt.Fprintf(w.writer, "HTTP/1.1 400 Bad Request\r\n")
	case 500:
		_, err = fmt.Fprintf(w.writer, "HTTP/1.1 500 Internal Server Error\r\n")
	default:
		_, err = fmt.Fprintf(w.writer, "HTTP/1.1 %d\r\n", statusCode)
	}

	if err == nil {
		w.state = stateStatusWritten
	}
	return err
}

// WriteHeaders writes HTTP header fields using the Writer's internal writer.
// Must be called after WriteStatusLine and before WriteBody.
// Each header follows RFC 9112 Section 3.2 format: field-name ":" field-value CRLF
func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != stateStatusWritten {
		return fmt.Errorf("WriteHeaders called out of order - must be called after WriteStatusLine")
	}

	for key, value := range headers {
		_, err := fmt.Fprintf(w.writer, "%s: %s\r\n", key, value)
		if err != nil {
			return err
		}
	}
	// Empty line marks end of headers section (RFC 9112 Section 3)
	_, err := fmt.Fprintf(w.writer, "\r\n")

	if err == nil {
		w.state = stateHeadersWritten
	}
	return err
}

// WriteBody writes raw []byte data to the response body using the Writer's internal writer.
// Must be called after WriteHeaders. Can be called multiple times.
// Returns the number of bytes written and any error encountered.
func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != stateHeadersWritten && w.state != stateBodyWritten {
		return 0, fmt.Errorf("WriteBody called out of order - must be called after WriteHeaders")
	}

	n, err := w.writer.Write(p)
	if err == nil {
		w.state = stateBodyWritten
	}
	return n, err
}

// WriteChunkedBody writes data as a chunked transfer encoding chunk.
// Must be called after WriteHeaders. Format: [hex-size]\r\n[data]\r\n
// Each call writes one complete chunk. Use WriteChunkedBodyDone() to finish.
func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != stateHeadersWritten && w.state != stateChunkedWriting {
		return 0, fmt.Errorf("WriteChunkedBody called out of order - must be called after WriteHeaders")
	}

	// Write chunk size in hexadecimal + CRLF
	chunkSize := len(p)
	_, err := fmt.Fprintf(w.writer, "%x\r\n", chunkSize)
	if err != nil {
		return 0, err
	}

	// Write chunk data
	n, err := w.writer.Write(p)
	if err != nil {
		return n, err
	}

	// Write trailing CRLF after chunk data
	_, err = fmt.Fprintf(w.writer, "\r\n")
	if err != nil {
		return n, err
	}

	w.state = stateChunkedWriting
	return n, nil
}

// WriteChunkedBodyDone writes the final chunk terminator for chunked encoding.
// Writes "0\r\n\r\n" to signal end of chunked response per RFC 9112 Section 7.1.3.
// Must be called after WriteChunkedBody to properly terminate the response.
func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != stateChunkedWriting {
		return 0, fmt.Errorf("WriteChunkedBodyDone called out of order - must be called after WriteChunkedBody")
	}

	// Write final chunk: size 0 + CRLF + CRLF (no trailing headers)
	n, err := fmt.Fprintf(w.writer, "0\r\n\r\n")
	if err == nil {
		w.state = stateChunkedDone
	}
	return n, err
}
