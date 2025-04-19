package main

import (
	"fmt"
	"net"
)

func main() {
	// Creating a Listening socket
	network := "tcp"
	address := "127.0.0.1:8080"
	listener, err := net.Listen(network, address)
	if err != nil {
		fmt.Printf("Error creating listening socket: %v\n", err)
	}
	defer listener.Close()
	fmt.Printf("Listening on %s://%s\n", network, listener.Addr())

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
	defer conn.Close()                          // Delays the execution of conn.Close() until handleConnection() has been properly executed
	fmt.Println(conn, "Hello from my goserver") // Sends the given data back to the connected client

	// Read data from the client

	buffer := make([]byte, 1024) // Message buffer
	n, err := conn.Read(buffer)  // conn.Read() reads the message in the buffer
	if err != nil {
		fmt.Println(err)
	}
	received := string(buffer[:n])
	fmt.Println(received)
	fmt.Fprintf(conn, "Thanks for your message: %s", received)
}
