package main

import (
	"log"
	"os/exec"

	"github.com/songgao/water"

	"l3vpn/internal/pf"
)

func main() {
	vpnGw := "10.0.0.3"

	tun, err := createTunInteface()
	if err != nil {
		panic(err)
	}

	log.Printf("Interface Name: %s\n", tun.Name())

	if err := setUpTunRouting(tun, vpnGw); err != nil {
		panic(err)
	}

	if err := setupPF(tun.Name(), vpnGw); err != nil {
		panic(err)
	}

	if err := listenTun(tun); err != nil {
		panic(err)
	}
}

func createTunInteface() (ifce *water.Interface, err error) {
	return water.New(water.Config{
		DeviceType: water.TUN,
	})
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

func listenTun(tun *water.Interface) error {

	packet := make([]byte, 2000)

	for {
		n, err := tun.Read(packet)
		if err != nil {
			return err
		}
		log.Printf("Packet Received: % x\n", packet[:n])
	}
}

func setupPF(tun, gateway string) error {
	pfConf := &pf.Config{
		Interface: tun,
		Gateway:   gateway,
	}
	return pfConf.ApplyRules()
}
