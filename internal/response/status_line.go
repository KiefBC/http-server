package response

import (
	"fmt"
	"io"
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
