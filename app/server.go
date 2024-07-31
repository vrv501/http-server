package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/vrv501/http-server/http"
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
		resp = http.CreateErrResponse(err)
		sendResponse(writer, resp)
		return
	}

	metadata := strings.Split(string(line), " ") // metadata is seperated by white-space
	if len(metadata) > 1 {                       // metadata should atleast have 2 strings
		path = metadata[1]
		method = metadata[0]
	} else {
		errMsg := "invalid metadata"
		resp = http.CreateHTTPResponse(400, map[string]string{
			http.ContentType:   http.PlainEncoding,
			http.ContentLength: fmt.Sprintf("%d", len(errMsg))},
			errMsg)
		sendResponse(writer, resp)
		return
	}

	// http method handling
	if method == "GET" {
		resp = http.HandleGETMethod(filesDir, path, reader)
	} else if method == "POST" {
		resp = http.HandlePOSTMethod(filesDir, path, reader)
	} else {
		errMsg := fmt.Sprintf("invalid method: %s", method)
		resp = http.CreateHTTPResponse(400, map[string]string{
			http.ContentType:   http.PlainEncoding,
			http.ContentLength: fmt.Sprintf("%d", len(errMsg))},
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
