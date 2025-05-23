package main

import (
	"bufio"
	"io"
	"log"
	"net"

	"l3vpn-server/internal/nat"
	"l3vpn-server/internal/tun"
	"l3vpn-server/internal/util"
	"l3vpn-server/protocol"
)

const port = "1337"

func main() {
	tun, err := tun.Create()
	if err != nil {
		log.Fatalf("failed to create TUN interface: %v", err)
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to start TCP listener: %v", err)
	}
	defer listener.Close()

	log.Printf("VPN server listening on port %s\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}

		log.Printf("client connected: %s", conn.RemoteAddr())

		nt := nat.NewNatTable()
		go handleConn(conn, tun, nt)
	}
}

func handleConn(conn net.Conn, tun *tun.TUN, nt *nat.NatTable) {
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

		packet, err := nat.TranslateOutbound(msg.Payload, nt, "127.0.0.2")
		if err != nil {
			log.Printf("failed to apply PAT: %v", err)
			continue
		}

		util.LogIPv4Packet(packet)

		if _, err := tun.Interface.Write(packet); err != nil {
			log.Printf("failed to write to TUN interface: %v", err)
			continue
		}
	}
}
