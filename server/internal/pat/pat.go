package pat

import (
	"errors"
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

const (
	Port = 1338
	IP   = "127.0.0.2"
)

func ChangeIPv4AndPort(originIPData []byte) ([]byte, error) {
	packet := gopacket.NewPacket(originIPData, layers.LayerTypeIPv4, gopacket.Default)

	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		return nil, errors.New("not an IP packet")
	}

	ip, ok := ipLayer.(*layers.IPv4)
	if !ok {
		return nil, errors.New("Failed to cast to IPv4 layer")
	}
	ip.SrcIP = net.ParseIP(IP).To4()

	var transport gopacket.SerializableLayer

	switch ip.Protocol {
	/*case layers.IPProtocolTCP:
		tcpLayer := packet.Layer(layers.LayerTypeTCP)
		if tcpLayer == nil {
			return nil, errors.New("not an TCP packet")
		}
		tcp, _ := tcpLayer.(*layers.TCP)
		tcp.SrcPort = layers.TCPPort(Port)
		tcp.SetNetworkLayerForChecksum(ip)
		transport = tcp
	case layers.IPProtocolUDP:
		udpLayer := packet.Layer(layers.LayerTypeUDP)
		if udpLayer == nil {
			return nil, errors.New("not an UDP packet")
		}
		udp, _ := udpLayer.(*layers.UDP)
		udp.SrcPort = layers.UDPPort(Port)
		udp.SetNetworkLayerForChecksum(ip)
		transport = udp*/
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

func encapsulate(layers ...gopacket.SerializableLayer) ([]byte, error) {
	buf := gopacket.NewSerializeBuffer()
	opt := gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
	if err := gopacket.SerializeLayers(buf, opt, layers...); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
