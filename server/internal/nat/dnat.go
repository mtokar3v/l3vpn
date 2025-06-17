package nat

import (
	"errors"
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func DNAT(originIPData []byte, nt *NatTable) ([]byte, error) {
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
		return dnatTCP(packet, ip, nt)
	case layers.IPProtocolUDP:
		return dnatUDP(packet, ip, nt)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", ip.Protocol)
	}
}

func dnatTCP(packet gopacket.Packet, ip *layers.IPv4, nt *NatTable) ([]byte, error) {
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	tcp, _ := tcpLayer.(*layers.TCP)

	srcSocket := Socket{IPAddr: ip.SrcIP.String(), Port: uint16(tcp.SrcPort)}
	dstSocket := Socket{IPAddr: ip.DstIP.String(), Port: uint16(tcp.DstPort)}

	tuple := FiveTuple{Src: dstSocket, Dst: srcSocket, Protocol: "TCP"}
	orgSocket, ok := nt.Get(tuple)
	if !ok {
		return nil, errors.New("unknown five tuple for dnat")
	}

	tcp.DstPort = layers.TCPPort(orgSocket.Port)
	ip.DstIP = net.ParseIP(orgSocket.IPAddr).To4()
	tcp.SetNetworkLayerForChecksum(ip)

	return serializePacket(ip, tcp, packet)
}

func dnatUDP(packet gopacket.Packet, ip *layers.IPv4, nt *NatTable) ([]byte, error) {
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	udp, _ := udpLayer.(*layers.UDP)

	srcSocket := Socket{IPAddr: ip.SrcIP.String(), Port: uint16(udp.SrcPort)}
	dstSocket := Socket{IPAddr: ip.DstIP.String(), Port: uint16(udp.DstPort)}

	tuple := FiveTuple{Src: dstSocket, Dst: srcSocket, Protocol: "UDP"}
	orgSocket, ok := nt.Get(tuple)
	if !ok {
		return nil, errors.New("unknown five tuple for dnat")
	}

	udp.DstPort = layers.UDPPort(orgSocket.Port)
	ip.DstIP = net.ParseIP(orgSocket.IPAddr).To4()
	udp.SetNetworkLayerForChecksum(ip)

	return serializePacket(ip, udp, packet)
}
