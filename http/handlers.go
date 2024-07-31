package http

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func HandleGETMethod(filesDir string, path string, reader *bufio.Reader) string {

	if path == "/" {
		return CreateHTTPResponse(200, map[string]string{}, "")
	} else if subPath, hasPrefix := strings.CutPrefix(path, "/echo/"); hasPrefix { // /echo/{str} should return str as response
		headers, err := readReqHeaders(reader)
		if err != nil {
			fmt.Println(err)
			return CreateErrResponse(err)
		}

		respHeaders := map[string]string{
			ContentType:   PlainEncoding,
			ContentLength: fmt.Sprintf("%d", len(subPath))}

		encodingSchemes := strings.Split(headers[AcceptEncoding], ",")
		for _, scheme := range encodingSchemes {
			if strings.TrimSpace(scheme) == "gzip" {
				respHeaders[ContentEncoding] = "gzip"
				break
			}
		}

		return CreateHTTPResponse(200, respHeaders, subPath)
	} else if strings.HasSuffix(path, "/user-agent") { // /user-agent should return header-value of User-Agent as response

		headers, err := readReqHeaders(reader)
		if err != nil {
			fmt.Println(err)
			return CreateErrResponse(err)
		}

		userAgent, ok := headers[UserAgent]
		if !ok {
			return CreateHTTPResponse(404, map[string]string{}, "")
		}

		return CreateHTTPResponse(200, map[string]string{
			ContentType:   PlainEncoding,
			ContentLength: fmt.Sprintf("%d", len(userAgent))},
			userAgent)
	} else if fileName, hasPrefix := strings.CutPrefix(path, "/files/"); hasPrefix { // read file and send its content as response
		file, err := os.ReadFile(filesDir + fileName)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return CreateHTTPResponse(404, map[string]string{}, "")
			}

			fmt.Println(err)
			return err.Error()
		}

		return CreateHTTPResponse(200, map[string]string{
			ContentType:   OctetStreamEncoding,
			ContentLength: fmt.Sprintf("%d", len(file))},
			string(file))
	}

	return CreateHTTPResponse(404, map[string]string{}, "")
}

func HandlePOSTMethod(filesDir string, path string, reader *bufio.Reader) string {

	// get the fileName from path
	fileName, hasPrefix := strings.CutPrefix(path, "/files/")
	if !hasPrefix {
		errStr := "invalidPath: " + path
		return CreateHTTPResponse(400, map[string]string{
			ContentType:   PlainEncoding,
			ContentLength: fmt.Sprintf("%d", len(errStr))},
			errStr)
	}

	// extract req headers
	headers, err := readReqHeaders(reader)
	if err != nil {
		fmt.Println(err)
		return CreateErrResponse(err)
	}

	// extract reqBody length from headers
	contentLengthStr, ok := headers[ContentLength]
	if !ok {
		errStr := "content-length not found"
		return CreateHTTPResponse(400, map[string]string{
			ContentType:   PlainEncoding,
			ContentLength: fmt.Sprintf("%d", len(errStr))},
			errStr)
	}

	contentLength, err := strconv.ParseUint(contentLengthStr, 10, 64)
	if err != nil {
		errStr := err.Error()
		return CreateHTTPResponse(400, map[string]string{
			ContentType:   PlainEncoding,
			ContentLength: fmt.Sprintf("%d", len(errStr))},
			errStr)
	}

	inpSlice := make([]byte, contentLength)
	_, err = reader.Read(inpSlice)
	if err != nil {
		fmt.Println(err)
		return CreateErrResponse(err)
	}

	// create file
	f, err := os.Create(filesDir + fileName)
	if err != nil {
		fmt.Println(err)
		return CreateErrResponse(err)
	}
	defer f.Close()

	// write reqBody to file
	_, err = f.Write(inpSlice)
	if err != nil {
		fmt.Println(err)
		return CreateErrResponse(err)
	}

	return CreateHTTPResponse(201, map[string]string{}, "")
}
