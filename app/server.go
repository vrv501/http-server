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
	//writer := bufio.NewWriter(conn)
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
	if req[1] == "/" {
		resp = "HTTP/1.1 200 OK\r\n\r\n"
	} else {
		resp = "HTTP/1.1 404 Not Found\r\n\r\n"
	}

	_, err = conn.Write([]byte(resp))
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		return
	}

}
