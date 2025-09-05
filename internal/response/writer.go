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
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
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
