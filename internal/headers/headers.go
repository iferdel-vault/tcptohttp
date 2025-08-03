package headers

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

const crlf = "\r\n"

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

func (h *Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		// consume the CRLF
		return len(crlf), true, nil
	}

	parts := bytes.SplitN(data[:idx], []byte(":"), 2)
	key := strings.ToLower(string(parts[0]))

	if key != strings.TrimRight(key, " ") {
		return 0, false, fmt.Errorf("invalid header name: %s", key)
	}
	if key == "" {
		return 0, false, fmt.Errorf("invalid header name: %s", key)
	}

	value := bytes.TrimSpace(parts[1])
	key = strings.TrimSpace(key)

	allowedSpecials := "!#$%&'*+-.^_`|~"
	for _, c := range key {
		if !unicode.IsDigit(c) && !unicode.IsLetter(c) && !strings.ContainsRune(allowedSpecials, c) {
			return 0, false, fmt.Errorf("header key contains unaccepted characters: %q", c)
		}
	}

	if len(strings.Split(key, " ")) > 1 {
		return 0, false, fmt.Errorf("malformed key on header with extra space internally in the key: %s", key)
	}

	err = h.Set(key, string(value))
	if err != nil {
		return 0, false, err
	}
	return idx + len(crlf), false, nil
}

func (h *Headers) Set(key, value string) error {
	key = strings.ToLower(key)
	if val, ok := (*h)[key]; ok {
		(*h)[key] = fmt.Sprintf("%s, %s", val, value)
		return nil
	}
	(*h)[key] = value
	return nil
}

func (h *Headers) Get(key string) (string, error) {
	key = strings.ToLower(key)
	if _, ok := (*h)[key]; !ok {
		return "", fmt.Errorf("header key %q does not exists", key)
	}
	return (*h)[key], nil
}
