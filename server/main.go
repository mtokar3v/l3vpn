package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"strconv"

	"l3vpn-server/internal/config"
	"l3vpn-server/internal/connection"
	"l3vpn-server/internal/nat"
	"l3vpn-server/internal/protocol"
	"l3vpn-server/internal/tun"
	"l3vpn-server/internal/util"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// TODO: super straitforward solution like step 1, step 2 etc
// each step will contains detailed comments
func main() {
	nt := nat.NewNatTable()
	cp := connection.NewConnectionPool()
	tun, err := tun.NewTUN()
	if err != nil {
		log.Fatalf("failed to create TUN interface: %v", err)
	}
	log.Printf("TUN interface created: %s", tun.Name)

	go listenClientTCPTraffic(nt, tun, cp)
	go listenExternalIPTraffic(nt, cp)

	select {}
}

func listenClientTCPTraffic(nt *nat.NatTable, tun *tun.TUN, cp *connection.ConnectionPool) {
	listener, err := net.Listen("tcp4", ":"+config.VPNPort)
	if err != nil {
		log.Fatalf("failed to start TCP listener: %v", err)
	}
	defer listener.Close()

	log.Printf("VPN server listening on port %s\n", config.VPNPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}

		log.Printf("client connected: %s", conn.RemoteAddr())

		cp.Set(conn.RemoteAddr().String(), conn)
		go handleClientConn(conn, tun, nt)
	}
}

func handleClientConn(conn net.Conn, tun *tun.TUN, nt *nat.NatTable) {
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

		publicSocket := publicSocket(conn)

		packet, err := nat.SNAT(msg.Payload, nt, publicSocket)
		if err != nil {
			log.Printf("failed to apply SNAT: %v", err)
			continue
		}

		util.LogIPv4Packet("[INBOUND]", packet)

		tun.Interface.Write(packet)
	}
}

func publicSocket(c net.Conn) *nat.Socket {
	addrSrt := c.RemoteAddr().String()
	orgAddr, portStr, _ := net.SplitHostPort(addrSrt)
	orgPort, _ := strconv.Atoi(portStr)

	return &nat.Socket{
		IPAddr: orgAddr,
		Port:   uint16(orgPort),
	}
}

func listenExternalIPTraffic(nt *nat.NatTable, cp *connection.ConnectionPool) {
	iface := "eth0" // change this to your network interface
	handle, err := pcap.OpenLive(iface, 65536, true, pcap.BlockForever)
	handle.SetBPFFilter("host 146.190.62.39")
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	log.Println("Listening for packets on", iface)

	for packet := range packetSource.Packets() {
		ethLayer := packet.Layer(layers.LayerTypeEthernet)
		if ethLayer == nil {
			log.Println("not an Ethernet packet")
			continue
		}

		eth, _ := ethLayer.(*layers.Ethernet)

		socket, packet, err := nat.DNAT(eth.Payload, nt)
		if err != nil {
			log.Printf("failed to apply PAT: %v", err)
			continue
		}

		util.LogIPv4Packet("[OUTBOUND]", packet)

		sendIPPacketToClient(socket, packet, cp)
	}
}

func sendIPPacketToClient(socket *nat.Socket, rawIP []byte, connections *connection.ConnectionPool) bool {
	clientAddr := socket.IPAddr + ":" + strconv.Itoa(int(socket.Port))
	conn, ok := connections.Get(clientAddr)
	if !ok {
		log.Printf("try to send ip packet to unknown connection %s", clientAddr)
		return false
	}

	_, err := conn.Write(rawIP)
	if err != nil {
		log.Printf("Failed to write to connection %s: %v", clientAddr, err)
		return false
	}

	return true
}
