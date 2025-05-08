package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"l3vpn-client/internal/pf"
	"l3vpn-client/internal/route"
	"l3vpn-client/internal/tun"
)

const (
	localIP = "10.0.0.1"
	Gateway = "10.0.0.3"

	VpnIP   = "127.0.0.1" // TODO: read from console or settings
	VpnPort = "1337"      //
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	conn, err := createTCPConnection()
	if err != nil {
		log.Fatalf("Fatal. Failed to create TCP connection: %v", err)
		panic(err)
	}
	defer conn.Close()

	tun, err := createTUNConnection()
	if err != nil {
		log.Fatalf("Fatal. Failed to create TUN interface: %v", err)
	}

	if err := setupRoute(localIP, Gateway, tun.Name); err != nil {
		panic(err)
	}

	pfConf, err := setupPF(tun.Name, Gateway)
	if err != nil {
		panic(err)
	}

	go func() {
		<-ctx.Done()
		stopVPN(pfConf)
	}()

	for {
		if err := tun.ForwardPackets(conn); err != nil {
			log.Printf("Warning. Something went wrong during tun forwarding: %v", err)
			continue
		}
	}
}

func createTCPConnection() (net.Conn, error) {
	conn, err := net.Dial("tcp", VpnIP+":"+VpnPort)
	if err != nil {
		return nil, err
	}

	log.Println("Connection created")
	return conn, nil
}

func createTUNConnection() (*tun.TUN, error) {
	tun, err := tun.Create()
	if err != nil {
		return nil, err
	}

	log.Printf("Interface created: %s", tun.Name)
	return tun, nil
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

func stopVPN(pfConf *pf.Config) {
	log.Println("Cleaning rules...")
	pfConf.RemoveRules()
	os.Exit(0)
}
