package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
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
			fmt.Println("Error accepting connection: ", err.Error())
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

	line, _, err = reader.ReadLine()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return
		}
		fmt.Println(err)
		return
	}

	req := strings.Split(string(line), " ")
	var resp string

	path := req[1]
	if path == "/" {
		resp = "HTTP/1.1 200 OK\r\n\r\n"
	} else if strings.HasPrefix(path, "/echo/") {
		subPath, _ := strings.CutPrefix(path, "/echo/")
		resp = fmt.Sprintf(
			"HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(subPath), subPath)
	} else if strings.HasSuffix(path, "/user-agent") {
		var (
			userAgent   string
			prefixFound bool
		)
		for {
			line, _, err = reader.ReadLine()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Println(err)
				return
			}

			userAgent, prefixFound = strings.CutPrefix(string(line), "User-Agent: ")
			if prefixFound {
				break
			}
		}
		resp = fmt.Sprintf(
			"HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
	} else {
		resp = "HTTP/1.1 404 Not Found\r\n\r\n"
	}

	_, err = writer.WriteString(resp)
	if err != nil {
		fmt.Println("Error writing stringResp: ", err.Error())
		return
	}

	err = writer.Flush()
	if err != nil {
		fmt.Println("Error writing to connection: ", err.Error())
		return
	}
}
