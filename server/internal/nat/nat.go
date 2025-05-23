package nat

import (
	"errors"
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// snat
func TranslateOutbound(originIPData []byte, nt *NatTable, srcIPAddr string) ([]byte, error) {
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
		return translateTCP(packet, ip, srcIPAddr, nt)
	case layers.IPProtocolUDP:
		return translateUDP(packet, ip, srcIPAddr, nt)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", ip.Protocol)
	}
}

func translateTCP(packet gopacket.Packet, ip *layers.IPv4, srcIPAddr string, nt *NatTable) ([]byte, error) {
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		return nil, errors.New("not an TCP packet")
	}
	tcp, _ := tcpLayer.(*layers.TCP)

	orgSocket := Socket{IPAddr: ip.SrcIP.String(), Port: uint16(tcp.SrcPort)}
	dstSocket := Socket{IPAddr: ip.DstIP.String(), Port: uint16(tcp.DstPort)}
	ft := FiveTuple{Src: orgSocket, Dst: dstSocket, Protocol: "TCP"}
	srcSocket := getOrCreateSrcSocket(ft, srcIPAddr, nt)

	tcp.SrcPort = layers.TCPPort(srcSocket.Port)
	ip.SrcIP = net.ParseIP(srcIPAddr).To4()
	tcp.SetNetworkLayerForChecksum(ip)

	return serializePacket(ip, tcp, packet)
}

func translateUDP(packet gopacket.Packet, ip *layers.IPv4, srcIPAddr string, nt *NatTable) ([]byte, error) {
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if udpLayer == nil {
		return nil, errors.New("not an UDP packet")
	}
	udp, _ := udpLayer.(*layers.UDP)

	orgSocket := Socket{IPAddr: ip.SrcIP.String(), Port: uint16(udp.SrcPort)}
	dstSocket := Socket{IPAddr: ip.DstIP.String(), Port: uint16(udp.DstPort)}
	ft := FiveTuple{Src: orgSocket, Dst: dstSocket, Protocol: "UDP"}
	srcSocket := getOrCreateSrcSocket(ft, srcIPAddr, nt)

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

func serializePacket(ip *layers.IPv4, transport gopacket.SerializableLayer, packet gopacket.Packet) ([]byte, error) {
	if app := packet.ApplicationLayer(); app != nil {
		payload := gopacket.Payload(app.Payload())
		return encapsulate(ip, transport, payload)
	}
	return encapsulate(ip, transport)
}

func encapsulate(layers ...gopacket.SerializableLayer) ([]byte, error) {
	buf := gopacket.NewSerializeBuffer()
	opt := gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
	if err := gopacket.SerializeLayers(buf, opt, layers...); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
