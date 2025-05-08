package tun

import (
	"github.com/songgao/water"
)

const (
	packetBufferSize = 2000
)

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
