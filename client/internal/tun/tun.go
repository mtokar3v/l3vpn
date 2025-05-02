package tun

import (
	"l3vpn/internal/util"

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

func (t *TUN) Listen() error {
	packet := make([]byte, packetBufferSize)

	for {
		n, err := t.Interface.Read(packet)
		if err != nil {
			return err
		}

		util.LogIPv4Packet(packet[:n])
	}
}
