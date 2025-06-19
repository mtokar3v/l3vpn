package tun

import (
	"github.com/songgao/water"
)

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

func (t *TUN) Close() error {
	return t.Interface.Close()
}
