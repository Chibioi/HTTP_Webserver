package response

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"time"
)

// HTTP response struct

type Response struct {
	Status      int
	ContentType string
	Body        []byte
	Header      map[string]string
	Cookies     []*http.Cookie
}

// Creating a Base response

func BaseResponse(status int) *Response {
	return &Response{
		Status: status,
		Header: make(map[string]string), // Basic structure of the HTTP response
	}
}

// Adding a Response header
func (r *Response) WithHeader(key, value string) *Response {
	r.Header[key] = value
	return r // Generates the header and their respective key-value pair
}

// WithBody sets response body and Content-Type
func (r *Response) WithBody(contentType string, body []byte) *Response {
	r.ContentType = contentType
	r.Body = body
	return r.WithHeader("Content-Length", fmt.Sprintf("%d", len(body))) // Creates the body of the response
}

// WithCookie adds a Set-Cookie header
func (r *Response) WithCookie(cookie *http.Cookie) *Response {
	r.Cookies = append(r.Cookies, cookie)
	return r // sets the cookie header in the response
}

// Write sends the response to the client
func (r *Response) Write(conn net.Conn) error { // This sends the constructed HTTP response over the network connection (net/http)
	var buf bytes.Buffer

	// Status line - WRITES THE HTTP STATUS TO THE BUFFER
	buf.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n", r.Status, http.StatusText(r.Status)))

	// Standard headers - WRITES THE DATE HEADER AND SETS THE CONNECTION HEADER TO CLOSE
	buf.WriteString("Server: MyGoWebServer/1.0\r\n")
	buf.WriteString("Date: " + time.Now().Format(time.RFC1123) + "\r\n")
	buf.WriteString("Connection: close\r\n")

	// Content-Type (if set) - WRITES THE CONTENT HEADER TO THE BUFFER
	if r.ContentType != "" {
		buf.WriteString("Content-Type: " + r.ContentType + "\r\n")
	}

	// Custom headers - ITERATES THROUGH THE r.Header AND SETS CUSTOM HEADERS TO THEIR KEY-VALUE PAIRS
	for k, v := range r.Header {
		buf.WriteString(k + ": " + v + "\r\n")
	}

	// Cookies - SETS A COOKIE-HEADER FOR EACH COOKIE BY ITERATING THROUGH THE r.Cookie slice
	for _, cookie := range r.Cookies {
		buf.WriteString("Set-Cookie: " + cookie.String() + "\r\n") // cookie.String() formats the cookie to the correct header value
	}

	// End of headers - WRITES AN EMPTY LINE TO INDICATE THE END OF THE HTTP HEADERS AND BEGINNING OF THE RESPONSE BODY
	buf.WriteString("\r\n")

	// Body (if exists) - WRITES THE BODY TO THE BUFFER
	if len(r.Body) > 0 {
		buf.Write(r.Body)
	}

	_, err := conn.Write(buf.Bytes()) // WRITES THE ENTIRE CONTENT OF THE BUFFER TO THE net.conn()
	return err
}
