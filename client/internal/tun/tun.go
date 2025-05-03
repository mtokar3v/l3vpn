package tun

import (
	"encoding/binary"
	"l3vpn-client/internal/util"
	"log"
	"net"

	"github.com/songgao/water"
)

const packetBufferSize = 2000 // in bytes

type TUN struct {
	Interface *water.Interface
	Name      string
}

func Create() (*TUN, error) {
	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		return nil, err
	}

	return &TUN{
		Interface: ifce,
		Name:      ifce.Name(),
	}, nil
}

func (t *TUN) ForwardPackets(conn net.Conn) error {
	packet := make([]byte, packetBufferSize)

	for {
		n, err := t.Interface.Read(packet)
		if err != nil {
			return err
		}

		util.LogIPv4Packet(packet[:n])

		length := make([]byte, 2)
		binary.BigEndian.PutUint16(length, uint16(n))

		_, err = conn.Write(append(length, packet[:n]...))
		if err != nil {
			log.Printf("Failed to write to TCP connection: %v", err)
			continue
		}
	}
}
