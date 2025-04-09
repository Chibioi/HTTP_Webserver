// package main

// import (
// 	"files/packages/src/Sockethandler"
// 	"log"
// 	"net"
// 	"testing"
// )

// func TCPtest(t *testing.T) {
// 	listener, err := net.Listen("Tcp", ":8080")
// 	if err != nil {
// 		log.Fatalf("Error: %v \n", err)
// 		return
// 	}

// 	// Accepts new connections
// 	for {
// 		conn, err := listener.Accept()
// 		if err != nil {
// 			log.Fatalf("Error: %v \n", err)
// 			continue
// 		}

// 		go handler.Handleconnections(conn)
// 	}

// }

package main

import (
	"files/packages/src/Sockethandler" // Replace with the actual path to your Sockethandler package
	"net"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	// Start the server in a goroutine
	go main()

	// Give the server some time to start listening
	time.Sleep(1 * time.Second)

	// Attempt to connect to the server
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Send a simple message to the server (if needed by your handler)
	_, err = conn.Write([]byte("test message"))
	if err != nil {
		t.Fatalf("Failed to write to server: %v", err)
	}

	// Optionally, read a response from the server (if your handler sends one)
	buf := make([]byte, 1024)
	_, err = conn.Read(buf)
	if err != nil {
		t.Fatalf("Failed to read from server: %v", err)
	}

	// Add more assertions based on your handler's behavior
	// For example, check the content of the response or verify that the handler performs certain actions.
}

// Mock the handler function to avoid real handler logic during test.
func TestHandleconnections(t *testing.T) {
	//Create a dummy listener and connection.
	ln, err := net.Listen("tcp", "127.0.0.1:0") //Use random port
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer ln.Close()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Errorf("Accept failed: %v", err)
			return
		}
		defer conn.Close()

		//Simulate that connection is handled.
		handler.Handleconnections(conn)
	}()

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	//Add further verification. For example, if Handleconnections writes data back, you can read it here.
	//Since the original HandleConnections depends on your logic, more test code will be needed
}
