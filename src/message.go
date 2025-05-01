package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
)

// A parsed HTTP request header

type BodyReader struct {
	// length int
	read io.Reader
}

type HTTPerror struct {
	StatusCode int
	Message    string
}

// dynamic sized buffer

type Dynbuff struct {
	data   bytes.Buffer
	length int
}

// appending data to dynamic buffer
func Buffpush(buff *Dynbuff, data *bytes.Buffer) { // This pushes data to the buffer
	Databytes := data.Bytes()            // databytes creates an empty buffer
	_, err := buff.data.Write(Databytes) // bytes.Buffer.Write returns two values (int, err)
	if err != nil {
		// Handle the error appropriately, perhaps return it
		fmt.Println("Error writing to buffer:", err)
		return
	}
	buff.length += len(Databytes) // buffer length increases as you push data to the buffer
}

func SplitMessage(buff *Dynbuff) ([]byte, bool) { // SplitMessage() processes the Dynbuff to extract complete messages separated by the newline character (\n)
	index := bytes.IndexByte(buff.data.Bytes()[:buff.length], '\n') // locate the first occurence (first index) of the newline char (\n)

	if index == -1 { // Incomplete message: newline not found
		return nil, false
	}

	message := make([]byte, index+1)  // Creates a new Buffer containing the message up to (and including) the newline
	_, err := buff.data.Read(message) // This advances the buffer read pointer
	if err != nil {
		return nil, false // should not happen if the buffer is within bounds
	}

	Buffpop(buff, index+1) // Pops the processed image

	return message, true
}

func Buffpop(buff *Dynbuff, length int) {
	if length <= 0 {
		return // returns nothing if the length value is 0 or a negative number
	}
	if length >= buff.length { // checks if the no. of bytes you want to remove is greater than or equal to the buffer's valid length
		buff.data.Reset() // Effectively clears the buffer
		buff.length = 0
		return
	}

	remaining := buff.data.Bytes()[length:] // This effectively represents the data that remains after you conceptually remove the first length bytes.
	newBuffer := bytes.NewBuffer(remaining) // This creates and initialize a new buffer from the remaining data
	buff.data = *newBuffer
	buff.length -= length // buff.length is updated by subtracting the number(len) of bytes from it
}

func (e *HTTPerror) Error() string {
	return fmt.Sprintf("status %d: %s", e.StatusCode, e.Message)
}

func StatusError() error {
	// some logic
	return &HTTPerror{StatusCode: http.StatusBadRequest, Message: "Unexpected EOF."}
}

func Serveclient(conn net.Conn) error {
	defer conn.Close()
	buff := &Dynbuff{data: bytes.Buffer{}, length: 0}

	for {
		message, found := SplitMessage(buff)
		if !found {
			data := make([]byte, 1024)
			n, err := conn.Read(data)
			if err != nil {
				if err != io.EOF {
					fmt.Println("Error reading from connection:", err)
					return err
				}
				fmt.Println("Client disconnected")
				return nil
			}
			if n > 0 {
				Buffpush(buff, bytes.NewBuffer(data[:n]))
				continue
			}
			fmt.Println("Client disconnected (no new data).")
			return nil
		}

		// Handle echo protocol commands first
		if bytes.Equal(message, []byte("quit\n")) {
			_, err := conn.Write([]byte("Bye.\n"))
			if err != nil {
				fmt.Println("Error writing to connection:", err)
			}
			return nil
		}

		// Try to parse as HTTP request
		requestReader := bufio.NewReader(bytes.NewReader(message))
		httpRequest, err := http.ReadRequest(requestReader)
		if err != nil {
			// If not HTTP, treat as echo protocol
			reply := append([]byte("Echo: "), message...)
			_, err := conn.Write(reply)
			if err != nil {
				fmt.Println("Error writing to connection:", err)
				return err
			}
			continue
		}

		// Handle HTTP request
		requestBody := ReadFromRequest(conn, buff, httpRequest)
		if requestBody == nil {
			fmt.Println("Failed to read request body")
			continue
		}

		resp, err := handleReq(httpRequest, io.NopCloser(requestBody.read))
		if err != nil {
			fmt.Println("Error handling request:", err)
			continue
		}

		err = writeHTTPResponse(conn, resp)
		if err != nil {
			fmt.Println("Error writing response:", err)
			return err
		}

		resp.Body.Close()
	}
}

// Body Reader from an HTTP request

func ReadFromRequest(conn net.Conn, buff *Dynbuff, req *http.Request) *BodyReader {
	ContentLen := req.ContentLength
	Chunked := req.TransferEncoding

	bodyAllowed := !(req.Method == "GET" || req.Method == "HEAD")
	bodyBuffer := &bytes.Buffer{} // Use a separate buffer for the request body

	if bodyAllowed {
		if ContentLen > 0 {
			// "content-length" is present
			n, err := io.CopyN(bodyBuffer, conn, ContentLen)
			if err != nil {
				fmt.Printf("Error reading content-length body: %v\n", err)
				// Consider returning an error or a nil BodyReader here
				return &BodyReader{read: nil}
			}
			if n < ContentLen {
				fmt.Printf("Warning: Read %d bytes, expected %d\n", n, ContentLen)
			}
		} else if len(Chunked) > 0 {
			isChunked := false
			for _, enc := range Chunked {
				if enc == "chunked" {
					isChunked = true
					break
				}
			}
			if isChunked {
				// Chunked encoding - you'll need to implement chunked decoding here
				fmt.Println("Chunked encoding detected, needs implementation") // COME BACK TO THIS
				// You would typically use a bufio.Reader on the conn and parse chunks
				err := StatusError()
				if err != nil {
					fmt.Println(err)
					if httpErr, ok := err.(*HTTPerror); ok {
						fmt.Printf("HTTP error: status code %d, message %s\n", httpErr.StatusCode, httpErr.Message)
					}
					return &BodyReader{read: nil}
				}
			} else {
				// No Content-Length and not chunked
				err := StatusError()
				if err != nil {
					fmt.Println(err)
					if httpErr, ok := err.(*HTTPerror); ok {
						fmt.Printf("HTTP error: status code %d, message %s\n", httpErr.StatusCode, httpErr.Message)
					}
					return &BodyReader{read: nil}
				}
				// Handle case with no Content-Length and not chunked (e.g., read until EOF)
				n, err := io.Copy(bodyBuffer, conn)
				if err != nil {
					fmt.Printf("Error reading until EOF: %v\n", err)
					return &BodyReader{read: nil}
				}
				fmt.Printf("Read %d bytes until EOF\n", n)
			}
		} else {
			// No Content-Length and not chunked for a method that allows a body
			// This might indicate an error or a client sending data without proper headers
			err := StatusError()
			if err != nil {
				fmt.Println(err)
				if httpErr, ok := err.(*HTTPerror); ok {
					fmt.Printf("HTTP error: status code %d, message %s\n", httpErr.StatusCode, httpErr.Message)
				}
				return &BodyReader{read: nil}
			}
			// Consider how to handle this case - perhaps read until EOF?
			n, err := io.Copy(bodyBuffer, conn)
			if err != nil {
				fmt.Printf("Error reading until EOF: %v\n", err)
				return &BodyReader{read: nil}
			}
			fmt.Printf("Read %d bytes until EOF (no headers)\n", n)
		}
	}

	bodyReader := &BodyReader{read: bodyBuffer} // Assuming Dynbuff has an underlying byte slice
	return bodyReader
}

func readerFromMemory(data []byte) io.ReadCloser {
	return io.NopCloser(bytes.NewBuffer(data)) // NopCloser returns a ReadCloser with a no-op Close method wrapping the provided Reader data.
}

func handleReq(req *http.Request, body io.ReadCloser) (*http.Response, error) {
	var respBody io.ReadCloser

	switch req.URL.Path {
	case "/echo":
		respBody = body // Directly use the request body for echo
	default:
		respBody = readerFromMemory([]byte("hello world.\n"))
		// It's important to close the original body if you're not using it
		if body != nil {
			err := body.Close()
			if err != nil {
				fmt.Println("Error closing request body:", err)
			}
		}
	}

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Server": []string{"my_first_http_server"},
		},
		Body: respBody,
	}

	return resp, nil
}

func newConn(conn net.Conn) {
	defer conn.Close()

	err := Serveclient(conn) // Assuming Serveclient now returns an error
	if err != nil {
		fmt.Fprintf(os.Stderr, "exception: %v\n", err)

		var httpErr *HTTPerror
		if errors.As(err, &httpErr) {
			// intended to send an error response
			resp := &http.Response{
				StatusCode: httpErr.StatusCode,                               // Corrected field name
				Header:     make(http.Header),                                // Initialize Header as a map
				Body:       readerFromMemory([]byte(httpErr.Message + "\n")), // Corrected field name
			}
			resp.Header.Set("Content-Type", "text/plain") // Set the Content-Type header

			// Manually write the HTTP response to the connection
			err := writeHTTPResponse(conn, resp)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error writing HTTP response: %v\n", err)
			}
		}
	}
}

func writeHTTPResponse(conn net.Conn, resp *http.Response) error {
	_, err := fmt.Fprintf(conn, "HTTP/1.1 %d %s\r\n", resp.StatusCode, http.StatusText(resp.StatusCode))
	if err != nil {
		return err
	}
	for key, values := range resp.Header {
		for _, value := range values {
			_, err := fmt.Fprintf(conn, "%s: %s\r\n", key, value)
			if err != nil {
				return err
			}
		}
	}
	_, err = fmt.Fprintf(conn, "\r\n") // End of headers
	if err != nil {
		return err
	}
	_, err = io.Copy(conn, resp.Body)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}
