package tun

import (
	"l3vpn/client/config"
	"l3vpn/shared/util"

	"github.com/songgao/water"
)

type Tun struct {
	Interface *water.Interface
	Name      string
}

func NewTun() (*Tun, error) {
	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		return nil, err
	}
	tun := &Tun{
		Interface: ifce,
		Name:      ifce.Name(),
	}
	if err := util.EnableInfe(tun.Name, config.TUNLocalIP, config.TUNGateway, config.MTU); err != nil {
		return nil, err
	}

	return tun, nil
}

func (t *Tun) Close() error {
	if err := util.DisableInfe(t.Name); err != nil {
		return err
	}
	return t.Interface.Close()
}
