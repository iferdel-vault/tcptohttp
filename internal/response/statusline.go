package response

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
