package protocol

import "encoding/binary"

// [2 bytes length][payload]
type Leet struct {
	Length  uint16
	Payload []byte
}

func Create(payload []byte) *Leet {
	length := uint16(2 + len(payload))
	return &Leet{
		Length:  length,
		Payload: payload,
	}
}

func (l *Leet) Serialize() []byte {
	buf := make([]byte, l.Length)
	binary.BigEndian.PutUint16(buf[0:2], l.Length)
	copy(buf[2:], l.Payload)
	return buf
}

func DeserializeLeet(data []byte) *Leet {
	if len(data) < 2 {
		return nil
	}
	length := binary.BigEndian.Uint16(data[0:2])
	payload := data[2:]
	return &Leet{
		Length:  length,
		Payload: payload,
	}
}
