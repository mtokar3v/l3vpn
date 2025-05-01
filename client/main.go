package main

import (
	"log"
	"os/exec"

	"github.com/songgao/water"

	"l3vpn/internal/pf"
	"l3vpn/internal/tun"
)

func main() {
	vpnGw := "10.0.0.3"

	tun, err := tun.Create()
	if err != nil {
		panic(err)
	}
	log.Printf("Interface created: %s", tun.Name)

	if err := setUpTunRouting(tun.Interface, vpnGw); err != nil {
		panic(err)
	}

	if err := setupPF(tun.Name, vpnGw); err != nil {
		panic(err)
	}

	if err := tun.Listen(); err != nil {
		panic(err)
	}
}

func setUpTunRouting(tun *water.Interface, vpnGw string) error {
	kernelIpv4 := "10.0.0.1"

	err := exec.Command("sudo", "ifconfig", tun.Name(), kernelIpv4, vpnGw, "up").Run()
	if err != nil {
		return err
	}

	exec.Command("route", "add", "default", vpnGw).Run()
	if err != nil {
		return err
	}

	return nil
}

func setupPF(tun, gateway string) error {
	pfConf := &pf.Config{
		Interface: tun,
		Gateway:   gateway,
	}
	return pfConf.ApplyRules()
}
