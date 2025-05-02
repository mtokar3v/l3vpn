package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"l3vpn/internal/pf"
	"l3vpn/internal/route"
	"l3vpn/internal/tun"
)

const (
	localIP = "10.0.0.1"
	Gateway = "10.0.0.3"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	tun, err := tun.Create()
	if err != nil {
		panic(err)
	}
	log.Printf("Interface created: %s", tun.Name)

	if err := setupRoute(localIP, Gateway, tun.Name); err != nil {
		panic(err)
	}

	pfCong, err := setupPF(tun.Name, Gateway)
	if err != nil {
		panic(err)
	}

	go func() {
		<-ctx.Done()
		log.Print("Cleaning rules...")
		pfCong.RemoveRules()
		os.Exit(0)
	}()

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

func setupPF(tun, gateway string) (*pf.Config, error) {
	pfSetup := &pf.Config{
		Interface: tun,
		Gateway:   gateway,
	}
	return pfSetup, pfSetup.ApplyRules()
}
