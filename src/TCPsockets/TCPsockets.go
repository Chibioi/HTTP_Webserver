package main

import (
	"files/packages/src/Sockethandler" // Correct your import path
	"fmt"
	"log"
	"net"
	"net/http"
)

func main() {
	// TCP Socket Server (Port 8081)
	go func() {
		listener, err := net.Listen("tcp", ":8081") // Use a different port
		if err != nil {
			log.Fatalf("TCP Server Error: %v", err)
			return
		}
		defer listener.Close()

		fmt.Println("TCP Server listening on :8081")

		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("TCP Accept Error: %v", err)
				continue
			}
			go handler.Handleconnections(conn)
		}
	}()

	// HTTP Server (Port 8080)
	http.HandleFunc("/parse", handler.Parsingheader)
	fmt.Println("HTTP Server listening on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("HTTP Server Error: %v", err)
	}
}
