package nat

import (
	"errors"
	"fmt"
	"l3vpn-server/internal/config"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func SNAT(originIPData []byte, nt *NatTable, publicOrgSocket *Socket) ([]byte, error) {
	packet := gopacket.NewPacket(originIPData, layers.LayerTypeIPv4, gopacket.Default)

	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		return nil, errors.New("not an IP packet")
	}

	ip, ok := ipLayer.(*layers.IPv4)
	if !ok {
		return nil, errors.New("failed to cast to IPv4 layer")
	}

	switch ip.Protocol {
	case layers.IPProtocolTCP:
		return snatTCP(packet, ip, publicOrgSocket, nt)
	case layers.IPProtocolUDP:
		return snatUDP(packet, ip, publicOrgSocket, nt)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", ip.Protocol)
	}
}

func snatTCP(packet gopacket.Packet, ip *layers.IPv4, publicOrgSocket *Socket, nt *NatTable) ([]byte, error) {
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	tcp, _ := tcpLayer.(*layers.TCP)

	orgSockets := &SocketPair{
		Public: *publicOrgSocket,
		Private: Socket{
			IPAddr: ip.SrcIP.String(),
			Port:   uint16(tcp.SrcPort),
		},
	}
	dstSocket := Socket{IPAddr: ip.DstIP.String(), Port: uint16(tcp.DstPort)}

	// server <- vpn (if you know origin and destination it gives you source )
	tuple := &FiveTuple{Src: orgSockets.Public, Dst: dstSocket, Protocol: "TCP"}
	srcSocket := getOrCreateSrcSocket(tuple, nt)

	// vpn -> client (if you know source and destination it gives you origin)
	translatedTuple := &FiveTuple{Src: *srcSocket, Dst: dstSocket, Protocol: "TCP"}
	nt.Set(translatedTuple, orgSockets)

	tcp.SrcPort = layers.TCPPort(srcSocket.Port)
	ip.SrcIP = net.ParseIP(config.VPNAddress).To4()
	tcp.SetNetworkLayerForChecksum(ip)

	return serializePacket(ip, tcp, packet)
}

func snatUDP(packet gopacket.Packet, ip *layers.IPv4, publicOrgSocket *Socket, nt *NatTable) ([]byte, error) {
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	udp, _ := udpLayer.(*layers.UDP)

	orgSockets := &SocketPair{
		Public: *publicOrgSocket,
		Private: Socket{
			IPAddr: ip.SrcIP.String(),
			Port:   uint16(udp.SrcPort),
		},
	}
	dstSocket := Socket{IPAddr: ip.DstIP.String(), Port: uint16(udp.DstPort)}

	// server <- vpn (if you know origin and destination it gives you source )
	tuple := &FiveTuple{Src: orgSockets.Public, Dst: dstSocket, Protocol: "UDP"}
	srcSocket := getOrCreateSrcSocket(tuple, nt)

	// vpn -> client (if you know source and destination it gives you origin)
	translatedTuple := &FiveTuple{Src: *srcSocket, Dst: dstSocket, Protocol: "UDP"}
	nt.Set(translatedTuple, orgSockets)

	udp.SrcPort = layers.UDPPort(srcSocket.Port)
	ip.SrcIP = net.ParseIP(config.VPNAddress).To4()
	udp.SetNetworkLayerForChecksum(ip)

	return serializePacket(ip, udp, packet)
}

func getOrCreateSrcSocket(ft *FiveTuple, nt *NatTable) *Socket {
	srcSockets, ok := nt.Get(ft)
	if !ok {
		port := nt.RentPort()
		srcSockets = &SocketPair{
			Public: Socket{IPAddr: config.VPNAddress, Port: port},
			// Private: Socket{IPAddr: config.VPNAddress, Port: port},
		}
		nt.Set(ft, srcSockets)
	}
	return &srcSockets.Public
}
