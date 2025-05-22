package nat

import (
	"errors"
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func TranslateOutbound(originIPData []byte, natTable *NatTable, srcIPAddr string) ([]byte, error) {
	packet := gopacket.NewPacket(originIPData, layers.LayerTypeIPv4, gopacket.Default)

	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		return nil, errors.New("not an IP packet")
	}

	ip, ok := ipLayer.(*layers.IPv4)
	if !ok {
		return nil, errors.New("Failed to cast to IPv4 layer")
	}

	var transport gopacket.SerializableLayer
	srcPort := natTable.RentPort()

	// TODO: Move to the generic func
	switch ip.Protocol {
	case layers.IPProtocolTCP:
		tcpLayer := packet.Layer(layers.LayerTypeTCP)
		if tcpLayer == nil {
			return nil, errors.New("not an TCP packet")
		}
		tcp, _ := tcpLayer.(*layers.TCP)

		orgSocket := Socket{IPAddr: ip.SrcIP.String(), Port: uint16(tcp.SrcPort)}
		dstSocket := Socket{IPAddr: ip.DstIP.String(), Port: uint16(tcp.DstPort)}
		ft := FiveTuple{Src: orgSocket, Dst: dstSocket, Protocol: "TCP"}
		srcSocket := getSrcSocket(ft, srcIPAddr, natTable)
		tcp.SrcPort = layers.TCPPort(srcSocket.Port)
		ip.SrcIP = net.ParseIP(srcIPAddr).To4()

		tcp.SetNetworkLayerForChecksum(ip)
		transport = tcp
	case layers.IPProtocolUDP:
		udpLayer := packet.Layer(layers.LayerTypeUDP)
		if udpLayer == nil {
			return nil, errors.New("not an UDP packet")
		}
		udp, _ := udpLayer.(*layers.UDP)
		udp.SrcPort = layers.UDPPort(srcPort)
		udp.SetNetworkLayerForChecksum(ip)
		transport = udp
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", ip.Protocol)
	}

	var err error
	var changedIPData []byte
	// TODO: do i rly need to serialize app if i already put ip
	if app := packet.ApplicationLayer(); app != nil {
		payload := gopacket.Payload(app.Payload())
		changedIPData, err = encapsulate(ip, transport, payload)
	} else {
		changedIPData, err = encapsulate(ip, transport)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to serialize packet: %w", err)
	}

	return changedIPData, nil
}

func getSrcSocket(ft FiveTuple, srcIPAddr string, nt *NatTable) Socket {
	srcSocket, ok := nt.Get(ft)
	if !ok {
		port := nt.RentPort()
		srcSocket := Socket{IPAddr: srcIPAddr, Port: port}
		nt.Set(ft, srcSocket)
	}
	return srcSocket
}

func encapsulate(layers ...gopacket.SerializableLayer) ([]byte, error) {
	buf := gopacket.NewSerializeBuffer()
	opt := gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
	if err := gopacket.SerializeLayers(buf, opt, layers...); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
