package vpn

import (
	"bufio"
	"io"
	"l3vpn-client/internal/config"
	"l3vpn-client/internal/protocol"
	"l3vpn-client/internal/util"
	"log"
	"net"
	"strconv"
)

func Listen() {
	portStr := strconv.Itoa(config.VPNPort)

	listener, err := net.Listen("tcp4", ":"+portStr)
	if err != nil {
		log.Fatalf("failed to start VPN listener on port %s: %v", portStr, err)
	}
	defer listener.Close()

	log.Printf("listening server's response on port %s\n", portStr)
	startListeningLoop(listener)
}

func startListeningLoop(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}

		log.Printf("client connected: %s", conn.RemoteAddr())

		handleClientConn(conn)
	}
}

func handleClientConn(conn net.Conn) {
	defer func() {
		log.Printf("closing connection: %s", conn.RemoteAddr())
		conn.Close()
	}()

	reader := bufio.NewReader(conn)

	for {
		msg, err := protocol.Read(reader)
		if err != nil {
			if err == io.EOF {
				log.Printf("client disconnected: %s", conn.RemoteAddr())
				return
			}
			log.Printf("failed to read protocol message: %v", err)
			continue
		}

		util.LogIPv4Packet(msg.Payload)
	}
}
