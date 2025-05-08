package main

import (
	"bufio"
	"l3vpn-server/internal/encapsulation"
	"l3vpn-server/internal/pat"
	"l3vpn-server/internal/tun"
	"l3vpn-server/internal/util"
	"l3vpn-server/protocol"
	"log"
	"net"
)

const (
	port = "1337"
)

func main() {
	tun, err := tun.Create()
	if err != nil {
		log.Fatalf("Fatal. failed to create tun infe: %v", err)
		panic(err)
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Fatal. failed to start listening: %v", err)
		panic(err)
	}
	defer listener.Close()
	log.Printf("Server is running on port %s\n", port)

	for {
		conn, err := listenConn(listener)
		if err != nil {
			log.Println("Failed to accept connection %w", err)
			continue
		}

		go handleConnection(conn, tun)
	}
}

func handleConnection(conn net.Conn, tun *tun.TUN) {
	defer closeConn(conn)

	log.Printf("New connection %s\n", conn.RemoteAddr())
	reader := bufio.NewReader(conn)

	for {
		vp, err := protocol.Read(reader)
		if err != nil {
			log.Printf("Failed to read leet: %v", err)
			continue
		}

		originIPData, err := encapsulation.DeEncapsulateTCPPacket(vp.Payload)
		if err != nil {
			log.Printf("Failed to de-encapsulate TCP packet: %v", err)
			continue
		}

		changedIPData, err := pat.ChangeIPv4AndPort(originIPData)
		if err != nil {
			log.Printf("Failed to make Port Address Translation: %v", err)
			continue
		}
		util.LogIPv4Packet(changedIPData)

		if _, err := tun.Interface.Write(changedIPData); err != nil {
			log.Printf("Failed to write into TUN infe: %v", err)
			continue
		}
	}
}

func listenConn(listener net.Listener) (net.Conn, error) {
	log.Println("Listening connections")
	conn, err := listener.Accept()
	log.Println("connection accepted")
	return conn, err
}

func closeConn(c net.Conn) {
	log.Printf("Closing connection %s\n", c.RemoteAddr())
	c.Close()
}
