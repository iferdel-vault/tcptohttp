package response

import (
	"fmt"
	"io"

	"github.com/iferdel-vault/tcptohttp/internal/headers"
)

type writerState int

const (
	WriterStateStatusLine writerState = iota
	WriterStateHeaders
	WriterStateBody
)

type Writer struct {
	writerState writerState
	conn        io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writerState: WriterStateStatusLine,
		conn:        w,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != WriterStateStatusLine {
		return fmt.Errorf("cannot write status line in state %d", w.writerState)
	}
	defer func() { w.writerState = WriterStateHeaders }()

	_, err := w.conn.Write(getStatusLine(statusCode))
	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState != WriterStateHeaders {
		return fmt.Errorf("cannot write headers in state %d", w.writerState)
	}
	defer func() { w.writerState = WriterStateBody }()

	for key, value := range headers {
		_, err := w.conn.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
		if err != nil {
			return err
		}
	}
	_, err := w.conn.Write([]byte("\r\n"))
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != WriterStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
	}
	defer func() { w.writerState = WriterStateStatusLine }()

	n, err := w.conn.Write(p)
	return n, err
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.writerState != WriterStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
	}
	chunkSize := len(p)

	nTotal := 0
	n, err := fmt.Fprintf(w.conn, "%x\r\n", chunkSize)
	if err != nil {
		return nTotal, nil
	}
	nTotal += n

	n, err = w.conn.Write(p)
	if err != nil {
		return nTotal, err
	}
	nTotal += n

	n, err = w.conn.Write([]byte("\r\n"))
	if err != nil {
		return nTotal, err
	}
	nTotal += n

	return nTotal, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.writerState != WriterStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
	}
	n, err := w.conn.Write([]byte("0\r\n\r\n"))
	if err != nil {
		return n, err
	}
	return n, nil
}
