package protocol

import (
	"encoding/binary"
	"io"
)

const (
	headerLen = 2
)

// [2 bytes length][payload]
type VPNProtocol struct {
	Length  uint16
	Payload []byte
}

func NewVPNProtocol(payload []byte) *VPNProtocol {
	length := uint16(headerLen + len(payload))
	return &VPNProtocol{
		Length:  length,
		Payload: payload,
	}
}

func (l *VPNProtocol) Serialize() []byte {
	buf := make([]byte, l.Length)
	binary.BigEndian.PutUint16(buf[0:headerLen], l.Length)
	copy(buf[headerLen:], l.Payload)
	return buf
}

func Deserialize(data []byte) *VPNProtocol {
	if len(data) < headerLen {
		return nil
	}
	length := binary.BigEndian.Uint16(data[0:headerLen])
	payload := data[headerLen:]
	return &VPNProtocol{
		Length:  length,
		Payload: payload,
	}
}

func Read(reader io.Reader) (*VPNProtocol, error) {
	header := make([]byte, headerLen)
	if _, err := io.ReadFull(reader, header); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint16(header)
	payloadLength := length - uint16(len(header))

	payload := make([]byte, payloadLength)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, err
	}

	return &VPNProtocol{
		Length:  length,
		Payload: payload,
	}, nil
}
