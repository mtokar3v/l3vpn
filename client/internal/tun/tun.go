package tun

import (
	"fmt"
	"l3vpn-client/internal/protocol"
	"l3vpn-client/internal/util"
	"net"

	"github.com/songgao/water"
)

const packetBufferSize = 2000 // in bytes

type TUN struct {
	Interface *water.Interface
	Name      string
}

func NewTUN() (*TUN, error) {
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
	buf := make([]byte, packetBufferSize)
	for {
		n, err := t.Interface.Read(buf)
		if err != nil {
			return err
		}
		packet := buf[:n]
		util.LogIPv4Packet(packet)
		vp := protocol.NewVPNProtocol(packet) // wrap tun traffic into custom protocol
		_, err = conn.Write(vp.Serialize())
		if err != nil {
			return fmt.Errorf("Failed to write to TCP connection: %v", err)
		}
	}
}
