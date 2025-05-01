package main

import (
	"log"

	"l3vpn/internal/pf"
	"l3vpn/internal/route"
	"l3vpn/internal/tun"
)

const (
	localIP = "10.0.0.1"
	Gateway = "10.0.0.3"
)

func main() {
	tun, err := tun.Create()
	if err != nil {
		panic(err)
	}
	log.Printf("Interface created: %s", tun.Name)

	if err := setupRoute(localIP, Gateway, tun.Name); err != nil {
		panic(err)
	}

	if err := setupPF(tun.Name, Gateway); err != nil {
		panic(err)
	}

	if err := tun.Listen(); err != nil {
		panic(err)
	}
}

func setupRoute(localIP, gateway, interfaceName string) error {
	routeConf := &route.Config{
		InterfaceName: interfaceName,
		LocalIP:       localIP,
		Gateway:       gateway,
	}
	return routeConf.Setup()
}

func setupPF(tun, gateway string) error {
	pfConf := &pf.Config{
		Interface: tun,
		Gateway:   gateway,
	}
	return pfConf.ApplyRules()
}
