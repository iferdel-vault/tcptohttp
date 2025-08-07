package response

import (
	"fmt"
	"io"
)

const (
	StatusOK                  StatusCode = 200 // RFC 9110, 15.3.1
	StatusBadRequest          StatusCode = 400 // RFC 9110, 15.5.1
	StatusInternalServerError StatusCode = 500 // RFC 9110, 15.6.1
)

type StatusCode int

var StatusCodeReasonPhrase = map[StatusCode]string{
	StatusOK:                  "HTTP/1.1 200 OK",
	StatusBadRequest:          "HTTP/1.1 400 Bad Request",
	StatusInternalServerError: "HTTP/1.1 500 Internal Server Error",
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	val, ok := StatusCodeReasonPhrase[statusCode]
	if !ok {
		val = fmt.Sprintf("HTTP/1.1 %d ", statusCode)
	}
	_, err := w.Write([]byte(val))
	return err
}
