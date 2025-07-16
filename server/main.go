package main

import (
	"io"
	"l3vpn/server/config"
	"l3vpn/server/connection"
	"l3vpn/server/nat"
	"l3vpn/shared/tun"
	"l3vpn/shared/util"
	"log"
	"net"
	"strconv"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// TODO: super straitforward solution like step 1, step 2 etc
// each step will contains detailed comments
func main() {
	nt := nat.NewNatTable()
	cp := connection.NewConnectionPool()
	tun, err := tun.NewTun()
	if err != nil {
		log.Fatalf("failed to create TUN interface: %v", err)
	}
	err = tun.Up()
	if err != nil {
		log.Fatalf("failed to up tun interface: %v", err)
	}
	log.Printf("TUN interface created: %s", tun.Name)
	go listenClientTCPTraffic(nt, tun, cp)
	go listenExternalIPTraffic(nt, cp)
	select {}
}

func listenClientTCPTraffic(nt *nat.NatTable, tun *tun.Tun, cp *connection.ConnectionPool) {
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
		tcpConn, ok := conn.(*net.TCPConn)
		if !ok {
			log.Printf("not a TCP connection")
			continue
		}

		log.Printf("client connected: %s", conn.RemoteAddr())
		cp.Set(conn.RemoteAddr().String(), tcpConn)
		go handleClientConn(tcpConn, tun, nt)
	}
}

func handleClientConn(conn *net.TCPConn, tun *tun.Tun, nt *nat.NatTable) {
	defer func() {
		log.Printf("closing connection: %s", conn.RemoteAddr())
		conn.Close()
	}()
	for {
		packet, err := util.ReadPacket(conn)
		if err != nil {
			if err == io.EOF {
				log.Printf("client disconnected: %s", conn.RemoteAddr())
				return
			}
			log.Printf("failed to read protocol message: %v", err)
			continue
		}
		publicSocket := publicSocket(conn)
		packet, err = nat.SNAT(packet, nt, publicSocket)
		if err != nil {
			log.Printf("failed to apply SNAT: %v", err)
			continue
		}
		util.LogIPv4Packet("[INBOUND]", packet)
		tun.Infe.Write(packet)
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
	handle.SetBPFFilter("src host 146.190.62.39")
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

func sendIPPacketToClient(socket *nat.Socket, ipPacket []byte, connections *connection.ConnectionPool) bool {
	clientAddr := socket.IPAddr + ":" + strconv.Itoa(int(socket.Port))
	conn, ok := connections.Get(clientAddr)
	if !ok {
		log.Printf("try to send ip packet to unknown connection %s", clientAddr)
		return false
	}
	_, err := util.WritePacket(conn, ipPacket)
	if err != nil {
		log.Printf("Failed to write to connection %s: %v", clientAddr, err)
		return false
	}
	return true
}
