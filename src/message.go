package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
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
func Buffpush(buff *Dynbuff, data *bytes.Buffer) {
	Databytes := data.Bytes()            // databytes creates an empty buffer
	_, err := buff.data.Write(Databytes) // bytes.Buffer.Write returns two values (int, err)
	if err != nil {
		// Handle the error appropriately, perhaps return it
		fmt.Println("Error writing to buffer:", err)
		return
	}
	buff.length += len(Databytes) // length of bytes of databytes is subtracted from buff.length
}

func SplitMessage(buff *Dynbuff) ([]byte, bool) { // SplitMessage() processes the Dynbuff to extract complete messages separated by the newline character (\n)
	index := bytes.IndexByte(buff.data.Bytes()[:buff.length], '\n') // locate the first occurence (first index) of the newline char (\n)
	if buff.length < 0 {
		return nil, false // incomplete message
	}

	message := make([]byte, index+1) // Creates a new Buffer containing the message up to (and including) the newline
	_, err := buff.data.Read(message)
	if err != nil {
		return nil, false
	}

	Buffpop(buff, index+1)

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

func Serveclient(conn net.Conn) {
	defer conn.Close() // Ensures the connection is close when the function exits
	buff := &Dynbuff{data: bytes.Buffer{}, length: 0}
	reader := bytes.NewReader(nil) // Initiates a new Reader function for the buffer

	for {
		// Trying to get one message from the buffer
		message, found := SplitMessage(buff)
		if !found {
			// Need more data
			data := make([]byte, 1024) // Buffer size is 1024
			length := len(data)
			n, err := conn.Read(data)
			if err != nil {
				if err != io.EOF {
					fmt.Println("Error reading from connection:", err)
					return // Connection closed or error occured
				}
			}
			if n > 0 {
				bytebuffer := bytes.NewBuffer(data[:n]) // creates a new buffer for the byte slice (from the beginning of the slice to n)
				Buffpush(buff, bytebuffer)
				if length == 0 && buff.length == 0 {
					return // no more requests
				}

				if length == 0 {
					err := StatusError()
					if err != nil {
						// Handle the error
						fmt.Println(err)
						// Check if the error is an HTTPError and handle accordingly
						if httpErr, ok := err.(*HTTPerror); ok {
							fmt.Printf("HTTP error: status code %d, message %s\n", httpErr.StatusCode, httpErr.Message)
						}
					}
				}

				// Update the reader to reflect the new buffer content
				reader.Reset(buff.data.Bytes()) // Resets the byte slice to be reading from "buff.data.Bytes()"
				continue                        // Get some more data and try again
			}
			fmt.Println("Client disconnected.")
			return
		}
		if bytes.Equal(message, []byte("quit\n")) { // checks if message and []byte("quit\n") is the same length and contain the same bytes
			_, err := conn.Write([]byte("Bye.\n")) // Writes "Bye" to the connection
			if err != nil {
				fmt.Println("Error writing to connection:", err)
			}
			return
		} else {
			reply := bytes.Join([][]byte{[]byte("Echo: "), message}, nil) // concatenates both slices
			_, err := conn.Write(reply)
			if err != nil {
				fmt.Println("Error writing to connection:", err)
			}
		}

		// Process the message and send the response
		Messages := bytes.NewBuffer(message)
		Request := io.NopCloser(Messages)

		requestBody := ReadFromRequest(conn, buff, Request)
	} // Loops end here
}

// Body Reader from an HTTP request

func ReadFromRequest(conn net.Conn, buff *Dynbuff, req *http.Request) *BodyReader {
	ContentLen := req.ContentLength
	Chunked := req.TransferEncoding

	bodyAllowed := !(req.Method == "GET" || req.Method == "HEAD")

	if bodyAllowed {
		if ContentLen > 0 {
			// "content-length" is present
			n, err := io.CopyN(&buff.data, conn, ContentLen)
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
				err := StatusError()
				if err != nil {
					fmt.Println(err)
					if httpErr, ok := err.(*HTTPerror); ok {
						fmt.Printf("HTTP error: status code %d, message %s\n", httpErr.StatusCode, httpErr.Message)
					}
					return &BodyReader{read: nil}
				}
				// Handle case with no Content-Length and not chunked (e.g., read until EOF)
				n, err := io.Copy(&buff.data, conn)
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
			n, err := io.Copy(&buff.data, conn)
			if err != nil {
				fmt.Printf("Error reading until EOF: %v\n", err)
				return &BodyReader{read: nil}
			}
			fmt.Printf("Read %d bytes until EOF (no headers)\n", n)
		}
	}

	bodyReader := &BodyReader{read: &buff.data} // Assuming Dynbuff has an underlying byte slice
	return bodyReader
}
