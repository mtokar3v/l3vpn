package tun

import (
	"l3vpn-client/internal/config"
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
	if err := network.EnableInfe(tun.Name, config.TUNLocalIP, config.TUNGateway, config.MTU); err != nil {
		return nil, err
	}

	return tun, nil
}

func (t *TUN) Close() error {
	if err := network.DisableInfe(t.Name); err != nil {
		return err
	}

	return t.Interface.Close()
}
