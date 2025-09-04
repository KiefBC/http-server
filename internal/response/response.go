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

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers.Set("content-length", fmt.Sprintf("%d", contentLen))
	headers.Set("connection", "close")
	headers.Set("content-type", "text/plain")
	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", key, value)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(w, "\r\n")
	return err
}
