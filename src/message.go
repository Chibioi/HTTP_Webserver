package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

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

	buff.length += len(Databytes)
	// const newLen = buff.length + databytes.length

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
			// conn.Close() at the defer will handle closing.
			return
		} else {
			reply := bytes.Join([][]byte{[]byte("Echo: "), message}, nil) // concatenates both slices
			_, err := conn.Write(reply)
			if err != nil {
				fmt.Println("Error writing to connection:", err)
			}
		}
	} // Loops end here
}
