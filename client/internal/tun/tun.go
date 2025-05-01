package tun

import (
	"log"

	"github.com/songgao/water"
)

const buffer = 2000

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
	packet := make([]byte, buffer)

	for {
		n, err := t.Interface.Read(packet)
		if err != nil {
			return err
		}
		log.Printf("Packet Received: % x\n", packet[:n])
	}
}
