package nat

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

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
