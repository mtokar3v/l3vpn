package nat

import (
	"errors"
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func SNAT(originIPData []byte, nt *NatTable, srcIPAddr string) ([]byte, error) {
	packet := gopacket.NewPacket(originIPData, layers.LayerTypeIPv4, gopacket.Default)

	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		return nil, errors.New("not an IP packet")
	}

	ip, ok := ipLayer.(*layers.IPv4)
	if !ok {
		return nil, errors.New("Failed to cast to IPv4 layer")
	}

	switch ip.Protocol {
	case layers.IPProtocolTCP:
		return snatTCP(packet, ip, srcIPAddr, nt)
	case layers.IPProtocolUDP:
		return snatUDP(packet, ip, srcIPAddr, nt)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", ip.Protocol)
	}
}

func snatTCP(packet gopacket.Packet, ip *layers.IPv4, srcIPAddr string, nt *NatTable) ([]byte, error) {
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	tcp, _ := tcpLayer.(*layers.TCP)

	orgSocket := Socket{IPAddr: ip.SrcIP.String(), Port: uint16(tcp.SrcPort)}
	dstSocket := Socket{IPAddr: ip.DstIP.String(), Port: uint16(tcp.DstPort)}

	// server <- vpn (if you know origin and destination it gives you source )
	tuple := FiveTuple{Src: orgSocket, Dst: dstSocket, Protocol: "TCP"}
	srcSocket := getOrCreateSrcSocket(tuple, srcIPAddr, nt)

	// vpn -> client (if you know source and destination it gives you origin)
	translatedTuple := FiveTuple{Src: srcSocket, Dst: dstSocket, Protocol: "TCP"}
	nt.Set(translatedTuple, orgSocket)

	tcp.SrcPort = layers.TCPPort(srcSocket.Port)
	ip.SrcIP = net.ParseIP(srcIPAddr).To4()
	tcp.SetNetworkLayerForChecksum(ip)

	return serializePacket(ip, tcp, packet)
}

func snatUDP(packet gopacket.Packet, ip *layers.IPv4, srcIPAddr string, nt *NatTable) ([]byte, error) {
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	udp, _ := udpLayer.(*layers.UDP)

	orgSocket := Socket{IPAddr: ip.SrcIP.String(), Port: uint16(udp.SrcPort)}
	dstSocket := Socket{IPAddr: ip.DstIP.String(), Port: uint16(udp.DstPort)}

	// server <- vpn (if you know origin and destination it gives you source )
	tuple := FiveTuple{Src: orgSocket, Dst: dstSocket, Protocol: "UDP"}
	srcSocket := getOrCreateSrcSocket(tuple, srcIPAddr, nt)

	// vpn -> client (if you know source and destination it gives you origin)
	translatedTuple := FiveTuple{Src: srcSocket, Dst: dstSocket, Protocol: "UDP"}
	nt.Set(translatedTuple, orgSocket)

	udp.SrcPort = layers.UDPPort(srcSocket.Port)
	ip.SrcIP = net.ParseIP(srcIPAddr).To4()
	udp.SetNetworkLayerForChecksum(ip)

	return serializePacket(ip, udp, packet)
}

func getOrCreateSrcSocket(ft FiveTuple, srcIPAddr string, nt *NatTable) Socket {
	srcSocket, ok := nt.Get(ft)
	if !ok {
		port := nt.RentPort()
		srcSocket = Socket{IPAddr: srcIPAddr, Port: port}
		nt.Set(ft, srcSocket)
	}
	return srcSocket
}
