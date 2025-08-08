package response

import (
	"fmt"
	"io"

	"github.com/iferdel-vault/tcptohttp/internal/headers"
)

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type StatusCode int

var StatusCodeReasonPhrase = map[StatusCode]string{
	StatusOK:                  "HTTP/1.1 200 OK",
	StatusBadRequest:          "HTTP/1.1 400 Bad Request",
	StatusInternalServerError: "HTTP/1.1 500 Internal Server Error",
}

type WriterState int

const (
	StatusLine WriterState = iota
	Headers
	Body
)

type Writer struct {
	conn        io.Writer
	writerState WriterState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		conn:        w,
		writerState: StatusLine,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != StatusLine {
		return fmt.Errorf("WriterState not expected before writting statusline: %d", w.writerState)
	}
	_, err := w.conn.Write(getStatusLine(statusCode))
	w.writerState = Headers
	return err
}
func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState != Headers {
		return fmt.Errorf("WriterState not expected before writting headers: %d", w.writerState)
	}
	for key, value := range headers {
		_, err := w.conn.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
		if err != nil {
			return err
		}
	}
	_, err := w.conn.Write([]byte("\r\n"))
	w.writerState = Body
	return err
}
func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != Body {
		return 0, fmt.Errorf("WriterState not expected before writting body: %d", w.writerState)
	}
	n, err := w.conn.Write(p)
	w.writerState = StatusLine
	return n, err
}

func getStatusLine(statusCode StatusCode) []byte {
	reasonPhrase := ""
	switch statusCode {
	case StatusOK:
		reasonPhrase = "OK"
	case StatusBadRequest:
		reasonPhrase = "Bad Request"
	case StatusInternalServerError:
		reasonPhrase = "Internal Server Error"
	}
	return []byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase))
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}
