package request

import (
	"errors"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	reqLine, err := ParseRequestLine(string(req))
	if err != nil {
		return nil, err
	}
	return &Request{
		RequestLine: reqLine,
	}, nil
}

func ParseRequestLine(req string) (RequestLine, error) {
	r := strings.Split(req, "\r\n")[0]

	rSlice := strings.Split(r, " ")
	if len(rSlice) != 3 {
		return RequestLine{}, errors.New("request line should have 3 entries separated by a space")
	}
	method, reqTarget, httpVersion := rSlice[0], rSlice[1], rSlice[2]

	if !IsUpper(method) {
		return RequestLine{}, errors.New("method should be all capital letters")
	}

	if !IsHTTPVersion(httpVersion, "1.1") {
		return RequestLine{}, errors.New("http version should be 1.1")
	}

	return RequestLine{
		HttpVersion:   strings.Split(httpVersion, "/")[1],
		RequestTarget: reqTarget,
		Method:        method,
	}, nil
}

func IsUpper(s string) bool {
	for _, r := range s {
		if !unicode.IsUpper(r) && unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func IsHTTPVersion(s, v string) bool {
	sv := strings.Split(s, "/")
	if sv[0] != "HTTP" {
		return false
	}
	return v == sv[1]
}
