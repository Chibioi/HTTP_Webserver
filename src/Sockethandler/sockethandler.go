package handler

import (
	"files/packages/src/Response_generation"
	"fmt"
	"html"
	"net"
	"net/http"
)

func Handleconnections(conn net.Conn) {
	// Closing the connection when we are done
	defer conn.Close()

	req := &response.Request{
		Path: "/api", // Example path
		Header: map[string]string{
			"Accept": "application/json",
		},
	}

	response.HandleRequest(conn, req)

	// Reading the incoming requests

	buff := make([]byte, 1024) // buffer that the messages will be stored
	_, err := conn.Read(buff)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Output the incoming requests
	fmt.Printf("Received: %v \n", buff)
	// Parsingheader()

}

// Parsing HTTP requests and Headers

func Parsingheader(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s %s %s \n", r.Method, r.URL, r.Proto)
	//Iterate over all header fields
	for k, v := range r.Header {
		fmt.Fprintf(w, "Header field %q, Value %q\n", k, v)
	}

	fmt.Fprintf(w, "Host = %q\n", r.Host)
	fmt.Fprintf(w, "RemoteAddr= %q\n", r.RemoteAddr)
	//Get value for a specified token
	fmt.Fprintf(w, "\n\nFinding value of \"Accept\" %q", r.Header["Accept"])

	// Handling HTTP methods

	switch r.Method {
	case "GET":
		fmt.Fprintf(w, "\nGET, %q", html.EscapeString(r.URL.Path))
	case "POST":
		fmt.Fprintf(w, "\nPOST, %q", html.EscapeString(r.URL.Path))
	case "PUT":
		fmt.Fprintf(w, "\nPUT, %q", html.EscapeString(r.URL.Path))
	case "DELETE":
		fmt.Fprintf(w, "\nDELETE, %q", html.EscapeString(r.URL.Path))
	default:
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}

}
