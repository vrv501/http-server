package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage

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
		line []byte
		err  error
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
		return
	}

	var (
		resp string
		path string
	)
	req := strings.Split(string(line), " ") // metadata is seperated by white-space
	if len(req) > 1 {
		path = req[1]
	}

	if path == "/" {
		resp = createHTTPResponse(200, map[string]string{}, "")
	} else if subPath, hasPrefix := strings.CutPrefix(path, "/echo/"); hasPrefix { // /echo/{str} should return str as response
		resp = createHTTPResponse(200, map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(subPath))},
			subPath)
	} else if strings.HasSuffix(path, "/user-agent") { // /user-agent should return header-value of User-Agent as response
		var (
			userAgent   string
			prefixFound bool
			lineStr     string
		)

		// each header is of format: key: value\r\n
		for {
			line, _, err = reader.ReadLine()
			if err != nil {
				fmt.Println(err)
				return
			}

			lineStr = string(line)
			if lineStr == "" { // when we have read all-lines, ReadLine returns empty-line
				break
			}
			userAgent, prefixFound = strings.CutPrefix(lineStr, "User-Agent: ")
			if prefixFound {
				break
			}
		}

		if userAgent == "" {
			resp = createHTTPResponse(404, map[string]string{}, "")
		} else {
			resp = createHTTPResponse(200, map[string]string{
				"Content-Type":   "text/plain",
				"Content-Length": fmt.Sprintf("%d", len(userAgent))},
				userAgent)
		}
	} else {
		resp = createHTTPResponse(404, map[string]string{}, "")
	}

	_, err = writer.WriteString(resp)
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

func createHTTPResponse(statusCode int, headers map[string]string, requestBody string) string {

	var headerStr strings.Builder
	for k, v := range headers {
		headerStr.WriteString(k)
		headerStr.WriteString(": ")
		headerStr.WriteString(v)
		headerStr.WriteString("\r\n")
	}

	var statusCodeStr string
	if statusCode == 200 {
		statusCodeStr = "OK"
	} else if statusCode == 404 {
		statusCodeStr = "Not Found"
	}

	return fmt.Sprintf("HTTP/1.1 %d %s\r\n%s\r\n%s", statusCode, statusCodeStr, headerStr.String(), requestBody)
}
