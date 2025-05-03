package main

import (
	"bufio"
	"encoding/binary"
	"io"
	"l3vpn-server/internal/util"
	"log"
	"net"
)

const (
	port = ":1337"
)

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	log.Printf("Server is running on port %s\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection %w", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	log.Printf("New connection %s\n", conn.RemoteAddr())
	reader := bufio.NewReader(conn)

	for {
		lenBytes := make([]byte, 2)
		if _, err := reader.Read(lenBytes); err != nil {
			log.Printf("Failed to read length: %v", err)
			return
		}
		length := binary.BigEndian.Uint16(lenBytes)

		packet := make([]byte, length)
		if _, err := io.ReadFull(reader, packet); err != nil {
			log.Printf("Failed to read packet: %v", err)
			return
		}

		util.LogIPv4Packet(packet)
	}
}
