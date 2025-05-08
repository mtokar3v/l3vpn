package encapsulation

import (
	"errors"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// de-encapsulates client's tcp segment to get original IP packet
func DeEncapsulateTCPPacket(data []byte) ([]byte, error) {
	packet := gopacket.NewPacket(data, layers.LayerTypeTCP, gopacket.Default)

	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		return nil, errors.New("Not an TCP packet")
	}

	tcp, ok := tcpLayer.(*layers.TCP)
	if !ok {
		return nil, errors.New("Failed to cast to TCP layer")
	}

	return tcp.Payload, nil
}
