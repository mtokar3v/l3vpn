package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"l3vpn-client/internal/forwarder"
	"l3vpn-client/internal/pf"
	"l3vpn-client/internal/route"
	"l3vpn-client/internal/tun"
)

const (
	localIP = "10.0.0.1"
	gateway = "10.0.0.3"
	vpnAddr = "127.0.0.1:1337" // TODO: move to config/args
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	conn, err := connectToVPN()
	if err != nil {
		log.Fatalf("failed to establish VPN connection: %v", err)
	}
	defer conn.Close()

	tun, err := tun.NewTUN()
	if err != nil {
		log.Fatalf("failed to create TUN interface: %v", err)
	}
	log.Printf("TUN interface created: %s", tun.Name)

	if err := route.Setup(&route.Config{
		Interface: tun.Name,
		LocalIP:   localIP,
		Gateway:   gateway,
	}); err != nil {
		log.Fatalf("route setup failed: %v", err)
	}

	port, err := strconv.Atoi(strings.Split(vpnAddr, ":")[1])
	pfConf := &pf.Config{
		Interface:  tun.Name,
		Gateway:    gateway,
		ByPassPort: port,
	}
	if err := pf.ApplyRules(pfConf); err != nil {
		log.Fatalf("pf setup failed: %v", err)
	}
	defer cleanup(pfConf)

	// Catch Ctrl+C or kill signal
	go func() {
		<-ctx.Done()
		cleanup(pfConf)
		os.Exit(0)
	}()

	for {
		if err := forwarder.ForwardPackets(tun, conn); err != nil {
			log.Printf("warning: packet forwarding error: %v", err)
			// reconnect logic could be placed here if desired
		}
	}
}

func connectToVPN() (net.Conn, error) {
	conn, err := net.Dial("tcp4", vpnAddr)
	if err != nil {
		return nil, err
	}
	log.Println("VPN connection established")
	return conn, nil
}

func cleanup(pfConf *pf.Config) {
	log.Println("Cleaning up PF rules...")
	if err := pf.RemoveRules(pfConf); err != nil {
		log.Printf("warning: failed to remove pf rules: %v", err)
	}
}
