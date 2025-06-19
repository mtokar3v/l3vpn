package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"syscall"

	"l3vpn-server/internal/connection"
	"l3vpn-server/internal/nat"
	"l3vpn-server/internal/util"
	"l3vpn-server/protocol"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

const (
	port     = "1337"
	publicIP = "147.93.120.166"
)

// TODO: super straitforward solution like step 1, step 2 etc
// each step will contains detailed comments
func main() {
	nt := nat.NewNatTable()
	go listenClientTCPTraffic(nt)
	go listenExternalIPTraffic(nt)

	select {}
}

func listenClientTCPTraffic(nt *nat.NatTable) {
	listener, err := net.Listen("tcp4", ":"+port)
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

		go handleClientConn(conn, nt)
	}
}

func handleClientConn(conn net.Conn, nt *nat.NatTable) {
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

		packet, err := nat.SNAT(msg.Payload, nt, publicIP)
		if err != nil {
			log.Printf("failed to apply SNAT: %v", err)
			continue
		}

		util.LogIPv4Packet(packet)

		sendIPPacket(packet)
	}
}

func sendIPPacket(rawIPPacket []byte) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	if err != nil {
		log.Printf("failed to create raw socket: %v", err)
		return
	}
	defer syscall.Close(fd)

	if err := syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1); err != nil {
		log.Printf("failed to set IP_HDRINCL: %v", err)
		return
	}

	var destIP [4]byte
	copy(destIP[:], rawIPPacket[16:20])
	log.Printf("Sending packet to IP: %v", destIP)

	dest := syscall.SockaddrInet4{
		Addr: destIP,
	}

	if err := syscall.Sendto(fd, rawIPPacket, 0, &dest); err != nil {
		log.Printf("syscall sendto failed: %v", err)
		return
	}
}

func listenExternalIPTraffic(nt *nat.NatTable) {
	iface := "eth0" // change this to your network interface
	handle, err := pcap.OpenLive(iface, 65536, true, pcap.BlockForever)
	handle.SetBPFFilter("not dst port " + port)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	connections := connection.NewConnectionPool()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	log.Println("Listening for packets on", iface)

	for packet := range packetSource.Packets() {

		ethLayer := packet.Layer(layers.LayerTypeEthernet)
		if ethLayer == nil {
			log.Println("not an Ethernet packet")
			return
		}

		eth, _ := ethLayer.(*layers.Ethernet)
		packet, err := nat.DNAT(eth.Payload, nt)
		if err != nil {
			log.Printf("failed to apply PAT: %v", err)
			continue
		}

		util.LogIPv4Packet(packet)

		sendIPPacketToClient(packet, connections)
	}
}

func sendIPPacketToClient(rawIP []byte, connections *connection.ConnectionPool) bool {
	packet := gopacket.NewPacket(rawIP, layers.LayerTypeIPv4, gopacket.Default)

	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		log.Printf("Not an IPv4 packet")
		return false
	}
	ip, _ := ipLayer.(*layers.IPv4)

	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if ipLayer == nil {
		log.Printf("Not an IPv4 packet")
		return false
	}
	tcp, _ := tcpLayer.(*layers.TCP)

	clientAddr := ip.DstIP.String() + ":" + tcp.DstPort.String()

	// Try to get existing connection
	conn, ok := connections.Get(clientAddr)
	if !ok {
		var err error
		conn, err = connectToClient(clientAddr)
		if err != nil {
			log.Printf("Failed to connect to client %s: %v", clientAddr, err)
			return false
		}
	}

	_, err := conn.Write(rawIP)
	if err != nil {
		log.Printf("Failed to write to connection %s: %v", clientAddr, err)
		return false
	}

	return true
}

func connectToClient(ipAddr string) (net.Conn, error) {
	conn, err := net.Dial("tcp4", ipAddr)
	if err != nil {
		return nil, err
	}
	log.Println("VPN connection established")
	return conn, nil
}
