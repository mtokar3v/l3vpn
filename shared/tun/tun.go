package tun

import (
	"l3vpn/client/config"
	"l3vpn/shared/util"

	"github.com/songgao/water"
)

type Tun struct {
	Infe *water.Interface
	Name string
}

func NewTun() (*Tun, error) {
	if ifce, err := water.New(water.Config{DeviceType: water.TUN}); err == nil {
		return &Tun{
			Infe: ifce,
			Name: ifce.Name(),
		}, nil
	} else {
		return nil, err
	}
}

func (t *Tun) Up() error {
	return util.EnableInfe(t.Name, config.TUNLocalIP, config.TUNGateway, config.MTU)
}

func (t *Tun) Close() error {
	if err := util.DisableInfe(t.Name); err != nil {
		return err
	}
	return t.Infe.Close()
}
