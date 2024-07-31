package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

var filesDir string

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	flag.StringVar(&filesDir, "directory", "", "directory name to search for files")
	flag.Parse()

	if filesDir != "" {
		os.Mkdir(filesDir, 0o777)
	}

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	var conn net.Conn
	for {
		conn, err = l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	defer func() { fmt.Println("done") }()

	fmt.Println("handling")

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	var (
		line   []byte
		err    error
		method string
		path   string
		resp   string
	)

	/*
		read request metdata. metadata has following info:
		http method
		path
		http version
	*/
	line, _, err = reader.ReadLine() //reads until \r\n or \n is encountered
	if err != nil {
		fmt.Println(err)
		resp = createErrResponse(err)
		sendResponse(writer, resp)
		return
	}

	metadata := strings.Split(string(line), " ") // metadata is seperated by white-space
	if len(metadata) > 1 {                       // metadata should atleast have 2 strings
		path = metadata[1]
		method = metadata[0]
	} else {
		errMsg := "invalid metadata"
		resp = createHTTPResponse(400, map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(errMsg))},
			errMsg)
		sendResponse(writer, resp)
		return
	}

	// http method handling
	if method == "GET" {
		resp = handleGETMethod(path, reader)
	} else if method == "POST" {
		resp = handlePOSTMethod(path, reader)
	} else {
		errMsg := fmt.Sprintf("invalid method: %s", method)
		resp = createHTTPResponse(400, map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(errMsg))},
			errMsg)
	}

	sendResponse(writer, resp)
}

func sendResponse(writer *bufio.Writer, resp string) { // sends response to client
	_, err := writer.WriteString(resp)
	if err != nil {
		fmt.Println("Error writing stringResp:", err.Error())
		return
	}

	err = writer.Flush()
	if err != nil {
		fmt.Println("Error writing to connection:", err.Error())
		return
	}
}

func createErrResponse(err error) string {

	errMsg := err.Error()
	return createHTTPResponse(500, map[string]string{
		"Content-Type":   "text/plain",
		"Content-Length": fmt.Sprintf("%d", len(errMsg))},
		errMsg)
}

func createHTTPResponse(statusCode int, headers map[string]string, requestBody string) string {

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

	return fmt.Sprintf("HTTP/1.1 %d %s\r\n%s\r\n%s", statusCode, statusCodeStr, headerStr.String(), requestBody)
}

func handlePOSTMethod(path string, reader *bufio.Reader) string {

	// get the fileName from path
	fileName, hasPrefix := strings.CutPrefix(path, "/files/")
	if !hasPrefix {
		errStr := "invalidPath: " + path
		return createHTTPResponse(400, map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(errStr))},
			errStr)
	}

	// extract req headers
	headers, err := readReqHeaders(reader)
	if err != nil {
		fmt.Println(err)
		return createErrResponse(err)
	}

	// extract reqBody length from headers
	contentLengthStr, ok := headers["Content-Length"]
	if !ok {
		errStr := "content-length not found"
		return createHTTPResponse(400, map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(errStr))},
			errStr)
	}

	contentLength, err := strconv.ParseUint(contentLengthStr, 10, 64)
	if err != nil {
		errStr := err.Error()
		return createHTTPResponse(400, map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(errStr))},
			errStr)
	}

	inpSlice := make([]byte, contentLength)
	_, err = reader.Read(inpSlice)
	if err != nil {
		fmt.Println(err)
		return createErrResponse(err)
	}

	// create file
	f, err := os.Create(filesDir + fileName)
	if err != nil {
		fmt.Println(err)
		return createErrResponse(err)
	}
	defer f.Close()

	// write reqBody to file
	_, err = f.Write(inpSlice)
	if err != nil {
		fmt.Println(err)
		return createErrResponse(err)
	}

	return createHTTPResponse(201, map[string]string{}, "")
}

func handleGETMethod(path string, reader *bufio.Reader) string {

	if path == "/" {
		return createHTTPResponse(200, map[string]string{}, "")
	} else if subPath, hasPrefix := strings.CutPrefix(path, "/echo/"); hasPrefix { // /echo/{str} should return str as response
		headers, err := readReqHeaders(reader)
		if err != nil {
			fmt.Println(err)
			return createErrResponse(err)
		}

		respHeaders := map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(subPath))}
		if headers["Accept-Encoding"] == "gzip" {
			respHeaders["Content-Encoding"] = "gzip"
		}

		return createHTTPResponse(200, respHeaders, subPath)
	} else if strings.HasSuffix(path, "/user-agent") { // /user-agent should return header-value of User-Agent as response

		headers, err := readReqHeaders(reader)
		if err != nil {
			fmt.Println(err)
			return createErrResponse(err)
		}

		userAgent, ok := headers["User-Agent"]
		if !ok {
			return createHTTPResponse(404, map[string]string{}, "")
		}

		return createHTTPResponse(200, map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(userAgent))},
			userAgent)
	} else if fileName, hasPrefix := strings.CutPrefix(path, "/files/"); hasPrefix { // read file and send its content as response
		file, err := os.ReadFile(filesDir + fileName)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return createHTTPResponse(404, map[string]string{}, "")
			}

			fmt.Println(err)
			return err.Error()
		}

		return createHTTPResponse(200, map[string]string{
			"Content-Type":   "application/octet-stream",
			"Content-Length": fmt.Sprintf("%d", len(file))},
			string(file))
	}

	return createHTTPResponse(404, map[string]string{}, "")
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
