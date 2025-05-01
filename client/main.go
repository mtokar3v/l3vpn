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

func setupRoute(localIP, gateway, ifce string) error {
	routeSetup := &route.Config{
		Interface: ifce,
		LocalIP:   localIP,
		Gateway:   gateway,
	}
	return routeSetup.Setup()
}

func setupPF(tun, gateway string) error {
	pfSetup := &pf.Config{
		Interface: tun,
		Gateway:   gateway,
	}
	return pfSetup.ApplyRules()
}
