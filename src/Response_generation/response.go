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
		Header: make(map[string]string),
	}
}

// Adding a Response header
func (r *Response) WithHeader(key, value string) *Response {
	r.Header[key] = value
	return r
}

// WithBody sets response body and Content-Type
func (r *Response) WithBody(contentType string, body []byte) *Response {
	r.ContentType = contentType
	r.Body = body
	return r.WithHeader("Content-Length", fmt.Sprintf("%d", len(body)))
}

// WithCookie adds a Set-Cookie header
func (r *Response) WithCookie(cookie *http.Cookie) *Response {
	r.Cookies = append(r.Cookies, cookie)
	return r
}

// Write sends the response to the client
func (r *Response) Write(conn net.Conn) error {
	var buf bytes.Buffer

	// Status line
	buf.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n", r.Status, http.StatusText(r.Status)))

	// Standard headers
	buf.WriteString("Date: " + time.Now().Format(time.RFC1123) + "\r\n")
	buf.WriteString("Connection: close\r\n")

	// Content-Type (if set)
	if r.ContentType != "" {
		buf.WriteString("Content-Type: " + r.ContentType + "\r\n")
	}

	// Custom headers
	for k, v := range r.Header {
		buf.WriteString(k + ": " + v + "\r\n")
	}

	// Cookies
	for _, cookie := range r.Cookies {
		buf.WriteString("Set-Cookie: " + cookie.String() + "\r\n")
	}

	// End of headers
	buf.WriteString("\r\n")

	// Body (if exists)
	if len(r.Body) > 0 {
		buf.Write(r.Body)
	}

	_, err := conn.Write(buf.Bytes())
	return err
}
