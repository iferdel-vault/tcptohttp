package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/iferdel-vault/tcptohttp/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	state       requestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type requestState int

const (
	requestStateInitialized requestState = iota
	requestStateDone
	requestStateParsingHeaders
)

const crlf = "\r\n"
const buffSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, buffSize, buffSize)
	readToIndex := 0
	req := &Request{
		state:   requestStateInitialized,
		Headers: headers.NewHeaders(),
	}
	for req.state != requestStateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				req.state = requestStateDone
				break
			}
			return nil, err
		}
		readToIndex += numBytesRead

		numBytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[numBytesParsed:])
		readToIndex -= numBytesParsed
	}
	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	var count int
	for r.state != requestStateDone {
		if r.state == requestStateInitialized {
			rl, i, err := parseRequestLine(data)
			if err != nil {
				return 0, err
			}
			if i == 0 {
				return count, nil
			}
			r.RequestLine = rl
			count += i
			r.state = requestStateParsingHeaders
		} else if r.state == requestStateParsingHeaders {
			n, done, err := r.Headers.Parse(data[count:])
			if err != nil {
				return 0, err
			}
			if n == 0 && !done {
				// not enough data for a new header yet
				return count, nil
			}
			count += n
			if done {
				r.state = requestStateDone
			}
		} else if r.state == requestStateDone {
			return 0, errors.New("error: trying to read data in a done state")
		} else {
			return 0, errors.New("error: unknown state")
		}
	}
	return count, nil
}

func parseRequestLine(data []byte) (RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return RequestLine{}, 0, nil
	}
	requestLineText := string(data[:idx])
	r, err := requestLineFromString(requestLineText)
	if err != nil {
		return RequestLine{}, 0, err
	}
	return *r, idx + len(crlf), nil
}

func requestLineFromString(str string) (*RequestLine, error) {
	parts := strings.Split(str, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("poorly formatted request-line: %s", str)
	}

	method := parts[0]
	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return nil, fmt.Errorf("invalid method: %s", method)
		}
	}

	requestTarget := parts[1]

	httpVersionParts := strings.Split(parts[2], "/")
	if len(httpVersionParts) != 2 {
		return nil, fmt.Errorf("malformed start-line: %s", str)
	}

	httpPart := httpVersionParts[0]
	if httpPart != "HTTP" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
	}
	version := httpVersionParts[1]
	if version != "1.1" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", version)
	}

	return &RequestLine{
		HttpVersion:   version,
		RequestTarget: requestTarget,
		Method:        method,
	}, nil

}
