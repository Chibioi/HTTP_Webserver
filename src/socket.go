// TCP echo server

package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	// Creating a Listening socket
	port := ":1234"
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
		go Serveclient(conn)
	}
}
