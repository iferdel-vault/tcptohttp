package response

import (
	"fmt"
	"io"

	"github.com/iferdel-vault/tcptohttp/internal/headers"
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

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.Headers{}
	h["Content-Length"] = fmt.Sprintf("%d", contentLen)
	h["Connection"] = "close"
	h["Content-Type"] = "text/plain"
	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		_, err := w.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
		if err != nil {
			return err
		}
	}
	return nil
}
