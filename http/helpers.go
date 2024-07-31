package http

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
)

func CreateErrResponse(err error) string {

	errMsg := err.Error()
	return CreateHTTPResponse(500, map[string]string{
		ContentType:   PlainEncoding,
		ContentLength: fmt.Sprintf("%d", len(errMsg))},
		[]byte(errMsg))
}

func CreateHTTPResponse(statusCode int, headers map[string]string, requestBody []byte) string {

	var headerStr strings.Builder
	for k, v := range headers {
		headerStr.WriteString(k)
		headerStr.WriteString(": ")
		headerStr.WriteString(v)
		headerStr.WriteString("\r\n")
	}

	var statusCodeStr string
	switch statusCode {
	case 200:
		statusCodeStr = "OK"
	case 201:
		statusCodeStr = "Created"
	case 400:
		statusCodeStr = "Bad Request"
	case 404:
		statusCodeStr = "Not Found"
	case 500:
		statusCodeStr = "Internal Server Error"
	}

	return fmt.Sprintf("HTTP/1.1 %d %s\r\n%s\r\n%v", statusCode, statusCodeStr, headerStr.String(), requestBody)
}

func readReqHeaders(reader *bufio.Reader) (map[string]string, error) {

	var (
		lineStr string
		headers = map[string]string{}
	)

	// each header is of format: key: value\r\n
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			return nil, err
		}

		lineStr = string(line)
		if lineStr == "" { // when we have read all-lines, ReadLine returns empty-line
			break
		}

		keyValue := strings.Split(lineStr, ": ")
		if len(keyValue) != 2 {
			return nil, errors.New("invalid header found")
		}

		headers[strings.TrimSpace(keyValue[0])] = strings.TrimSpace(keyValue[1])
	}

	return headers, nil
}
