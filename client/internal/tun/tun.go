package tun

import (
	"l3vpn-client/internal/network"

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
	tun := &TUN{
		Interface: ifce,
		Name:      ifce.Name(),
	}
	if err := network.Enable(tun.Name); err != nil {
		return nil, err
	}

	return tun, nil
}

func (t *TUN) Close() error {
	if err := network.Enable(t.Name); err != nil {
		return err
	}
	return t.Interface.Close()
}
