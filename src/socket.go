// TCP echo server

package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	// Creating a Listening socket
	port := ":8080"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Printf("Error creating listening socket: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Printf("Listening on %s:%s\n", "tcp", listener.Addr())

	// Accepting new connections on the listening socket
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error creating a new connection: %v\n", err)
			continue // Continue listening for more connections
		}

		fmt.Printf("Accepted connection from: %s\n", conn.RemoteAddr())
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close() // Delays the execution of conn.Close() until handleConnection() has been properly executed. conn.Close() closes the connection
	reader := bufio.NewReader(conn)
	for {
		// read client request data
		bytes, err := reader.ReadBytes(byte('\n'))
		if err != nil {
			if err != io.EOF {
				fmt.Println("failed to read data, err:", err)
			}
			return
		}
		fmt.Printf("request: %s", bytes)

		_, writeErr := conn.Write(bytes)
		if writeErr != nil {
			fmt.Println("failed to write response, err:", writeErr)
			return // Or handle the error appropriately
		}
		fmt.Printf("Response: %v", bytes)
	}
}
