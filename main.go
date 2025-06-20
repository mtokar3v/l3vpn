package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	serverAddr := "147.93.120.166:1337" // Replace with your server IP and port

	// Establish TCP connection
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to", serverAddr)

	// Send a single message
	message := "hello\n"
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send message: %v\n", err)
		return
	}

	fmt.Println("Message sent. Closing connection.")
	time.Sleep(5 * time.Second)
}
