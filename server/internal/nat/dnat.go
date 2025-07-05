package nat

import (
	"errors"
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func DNAT(originIPData []byte, nt *NatTable) (publicSocket *Socket, data []byte, err error) {
	packet := gopacket.NewPacket(originIPData, layers.LayerTypeIPv4, gopacket.Default)

	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		return nil, nil, errors.New("not an IP packet")
	}

	ip, ok := ipLayer.(*layers.IPv4)
	if !ok {
		return nil, nil, errors.New("failed to cast to IPv4 layer")
	}

	switch ip.Protocol {
	case layers.IPProtocolTCP:
		return dnatTCP(packet, ip, nt)
	case layers.IPProtocolUDP:
		return dnatUDP(packet, ip, nt)
	default:
		return nil, nil, fmt.Errorf("unsupported protocol: %s", ip.Protocol)
	}
}

func dnatTCP(packet gopacket.Packet, ip *layers.IPv4, nt *NatTable) (*Socket, []byte, error) {
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	tcp, _ := tcpLayer.(*layers.TCP)

	srcSocket := Socket{IPAddr: ip.SrcIP.String(), Port: uint16(tcp.SrcPort)}
	dstSocket := Socket{IPAddr: ip.DstIP.String(), Port: uint16(tcp.DstPort)}

	tuple := &FiveTuple{Src: dstSocket, Dst: srcSocket, Protocol: "TCP"}
	orgSockets, ok := nt.Get(tuple)
	if !ok {
		return nil, nil, fmt.Errorf(
			"unknown five tuple for dnat; saddr: %s:%d; daddr: %s:%d",
			tuple.Src.IPAddr, tuple.Src.Port,
			tuple.Dst.IPAddr, tuple.Dst.Port,
		)
	}

	tcp.DstPort = layers.TCPPort(orgSockets.Private.Port)
	ip.DstIP = net.ParseIP(orgSockets.Private.IPAddr).To4()
	tcp.SetNetworkLayerForChecksum(ip)

	data, err := serializePacket(ip, tcp, packet)
	return &orgSockets.Public, data, err
}

func dnatUDP(packet gopacket.Packet, ip *layers.IPv4, nt *NatTable) (*Socket, []byte, error) {
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	udp, _ := udpLayer.(*layers.UDP)

	srcSocket := Socket{IPAddr: ip.SrcIP.String(), Port: uint16(udp.SrcPort)}
	dstSocket := Socket{IPAddr: ip.DstIP.String(), Port: uint16(udp.DstPort)}

	tuple := &FiveTuple{Src: dstSocket, Dst: srcSocket, Protocol: "UDP"}
	orgSockets, ok := nt.Get(tuple)
	if !ok {
		return nil, nil, fmt.Errorf(
			"unknown five tuple for dnat; saddr: %s:%d; daddr: %s:%d",
			tuple.Src.IPAddr, tuple.Src.Port,
			tuple.Dst.IPAddr, tuple.Dst.Port,
		)
	}

	udp.DstPort = layers.UDPPort(orgSockets.Private.Port)
	ip.DstIP = net.ParseIP(orgSockets.Private.IPAddr).To4()
	udp.SetNetworkLayerForChecksum(ip)

	data, err := serializePacket(ip, udp, packet)
	return &orgSockets.Public, data, err
}
