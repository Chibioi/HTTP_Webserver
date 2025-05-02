package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"syscall"
	"time"
)

// A parsed HTTP request header

type BodyReader struct {
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
	Databytes := data.Bytes()
	_, err := buff.data.Write(Databytes)
	if err != nil {
		fmt.Println("Error writing to buffer:", err)
		return
	}
	buff.length += len(Databytes)
}

func SplitMessage(buff *Dynbuff) ([]byte, bool) {
	index := bytes.IndexByte(buff.data.Bytes()[:buff.length], '\n')

	if index == -1 {
		return nil, false // Incomplete message: newline not found
	}

	message := make([]byte, index+1)
	_, err := buff.data.Read(message)
	if err != nil {
		return nil, false // Should not happen if the buffer is within bounds
	}

	Buffpop(buff, index+1)

	return message, true
}

func Buffpop(buff *Dynbuff, length int) {
	if length <= 0 {
		return
	}
	if length >= buff.length {
		buff.data.Reset()
		buff.length = 0
		return
	}

	remaining := buff.data.Bytes()[length:]
	copy(buff.data.Bytes(), remaining) // Copy remaining data to the beginning
	buff.length -= length
	buff.data.Truncate(buff.length) // Truncate the buffer to the new length
}

func (e *HTTPerror) Error() string {
	return fmt.Sprintf("status %d: %s", e.StatusCode, e.Message)
}

func StatusError() error {
	return &HTTPerror{StatusCode: http.StatusBadRequest, Message: "Unexpected EOF."}
}

func Serveclient(conn net.Conn) {
	defer conn.Close()
	buff := &Dynbuff{data: bytes.Buffer{}, length: 0}
	httpHeaderReadTimeout := 5 * time.Second
	headerTimer := time.NewTimer(httpHeaderReadTimeout)
	defer headerTimer.Stop()
	isHTTPRequest := false

	for {
		select {
		case <-headerTimer.C:
			if !isHTTPRequest && buff.length > 0 {
				fmt.Println("Timeout reading HTTP headers, treating as simple message.")
				// Proceed to process as a non-HTTP message
			}
		default:
			data := make([]byte, 1024)
			err := conn.SetReadDeadline(time.Now().Add(1 * time.Second)) // Small timeout for reading chunks
			if err != nil {
				fmt.Println("Error setting read deadline:", err)
				return
			}
			n, err := conn.Read(data)
			if err != nil {
				if errors.Is(err, io.EOF) {
					fmt.Println("Client disconnected.")
					return
				}
				netErr, ok := err.(net.Error)
				if ok && netErr.Timeout() {
					// No data received within the short timeout, continue to check for headers or process existing buffer
				} else {
					fmt.Println("Error reading from connection:", err)
					return
				}
			}
			if n > 0 {
				bytebuffer := bytes.NewBuffer(data[:n])
				Buffpush(buff, bytebuffer)
				if !isHTTPRequest {
					// Check if we have the end of HTTP headers
					if bytes.Contains(buff.data.Bytes()[:buff.length], []byte("\r\n\r\n")) {
						isHTTPRequest = true
						headerTimer.Stop() // Stop the header read timeout
					}
				}
			}
		}

		if isHTTPRequest {
			// Process HTTP request
			requestReader := bufio.NewReader(bytes.NewReader(buff.data.Bytes()[:buff.length]))
			httpRequest, err := http.ReadRequest(requestReader)
			if err == nil {
				fmt.Println("Successfully parsed HTTP request.")
				// Consume the headers from the buffer
				headerEndIndex := bytes.Index(buff.data.Bytes()[:buff.length], []byte("\r\n\r\n"))
				if headerEndIndex != -1 {
					headerEnd := headerEndIndex + 4 // Include the \r\n\r\n
					Buffpop(buff, headerEnd)

					requestBody := ReadFromRequest(conn, httpRequest, buff)
					if requestBody == nil {
						fmt.Println("Failed to read request body (nil reader)")
						continue
					}
					defer func() {
						if closer, ok := requestBody.read.(io.Closer); ok {
							closer.Close()
						}
					}()

					resp, err := handleReq(httpRequest, requestBody.read)
					if err != nil {
						fmt.Println("Error handling request:", err)
						// Consider sending an error response
						continue
					}

					err = writeHTTPResponse(conn, resp)
					if err != nil {
						if ne, ok := err.(net.Error); ok && ne.Timeout() {
							fmt.Println("Write timeout:", err)
							return
						} else if errors.Is(err, syscall.EPIPE) || errors.Is(err, io.ErrClosedPipe) {
							fmt.Println("Error writing response (broken pipe):", err)
							return
						} else {
							fmt.Println("Error writing response:", err)
							return
						}
					}
					resp.Body.Close()
					return // Handle one HTTP request per connection for simplicity in this debugged version
				}
			} else {
				fmt.Printf("Error parsing HTTP request: %v\n", err)
				// Consider sending an error response
				return
			}
		} else {
			// Handle as simple line-based message (echo protocol)
			message, found := SplitMessage(buff)
			if found {
				if bytes.Equal(message, []byte("quit\n")) {
					_, err := conn.Write([]byte("Bye.\n"))
					if err != nil {
						fmt.Println("Error writing to connection:", err)
					}
					return
				}
				reply := append([]byte("Echo: "), message...)
				_, err := conn.Write(reply)
				if err != nil {
					if ne, ok := err.(net.Error); ok && ne.Timeout() {
						fmt.Println("Write timeout:", err)
						return
					} else if errors.Is(err, syscall.EPIPE) || errors.Is(err, io.ErrClosedPipe) {
						fmt.Println("Error writing echo response (broken pipe):", err)
						return
					} else {
						fmt.Println("Error writing echo response:", err)
						return
					}
				}
			}
		}

		if buff.length == 0 && isHTTPRequest {
			return // Connection likely closed after serving HTTP request
		}
	}
}

// Body Reader from an HTTP request
func ReadFromRequest(conn net.Conn, req *http.Request, buff *Dynbuff) *BodyReader {
	ContentLen := req.ContentLength
	Chunked := req.TransferEncoding
	bodyBuffer := &bytes.Buffer{}

	if !(req.Method == "GET" || req.Method == "HEAD") {
		if ContentLen > 0 {
			remaining := int64(buff.length)
			if remaining < ContentLen {
				bytesToRead := ContentLen - remaining
				bodyData := make([]byte, bytesToRead)
				n, err := io.ReadFull(conn, bodyData)
				if err != nil {
					fmt.Printf("Error reading full content-length body: %v\n", err)
					return &BodyReader{read: nil}
				}
				bodyBuffer.Write(buff.data.Bytes()[:buff.length])
				bodyBuffer.Write(bodyData[:n])
				Buffpop(buff, buff.length)
			} else {
				bodyBuffer.Write(buff.data.Bytes()[:ContentLen])
				Buffpop(buff, int(ContentLen))
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
				fmt.Println("Chunked encoding detected, needs full implementation")
				// In a real server, you would implement chunked decoding here.
				// For this debugged version, we'll not read a body in this case.
			} else {
				// No Content-Length and not chunked
				fmt.Println("No Content-Length and not chunked Transfer-Encoding, reading until EOF (might block)")
				_, err := bodyBuffer.ReadFrom(conn) // Potentially blocks until connection closes
				if err != nil {
					fmt.Printf("Error reading until EOF: %v\n", err)
					return &BodyReader{read: nil}
				}
				Buffpop(buff, buff.length)
			}
		} else {
			// No body expected or no way to determine length
			Buffpop(buff, buff.length)
		}
	} else {
		Buffpop(buff, buff.length) // Consume headers for GET/HEAD
	}

	return &BodyReader{read: bodyBuffer}
}

func readerFromMemory(data []byte) io.ReadCloser {
	return io.NopCloser(bytes.NewBuffer(data))
}

func handleReq(req *http.Request, body io.Reader) (*http.Response, error) {
	var respBody io.ReadCloser
	header := http.Header{}

	switch req.URL.Path {
	case "/echo":
		respBytes, err := io.ReadAll(body)
		if err != nil {
			fmt.Println("Error reading echo body:", err)
			return nil, errors.New("failed to read echo body")
		}
		respBody = readerFromMemory(respBytes)
		header.Set("Content-Length", fmt.Sprintf("%d", len(respBytes)))
		header.Set("Content-Type", "text/plain")
	default:
		response := []byte("hello world.\n")
		respBody = readerFromMemory(response)
		header.Set("Content-Length", fmt.Sprintf("%d", len(response)))
		header.Set("Content-Type", "text/plain")
	}

	header.Set("Server", "my_first_http_server")
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       respBody,
		Proto:      "HTTP/1.1", // Explicitly set the protocol (try this)
		ProtoMajor: 1,          // DEBUGGING ISSUES AROSE FROM HERE
		ProtoMinor: 1,
	}

	return resp, nil
}
func writeHTTPResponse(conn net.Conn, resp *http.Response) error {
	err := resp.Write(conn)
	return err
}
