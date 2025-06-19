package vpn

import (
	"context"
	"l3vpn-client/internal/config"
	"l3vpn-client/internal/pf"
	"l3vpn-client/internal/route"
	"l3vpn-client/internal/tools"
	"l3vpn-client/internal/tun"
	"log"
	"net"
	"os"
	"strconv"
)

func Forward(ctx context.Context) {
	tcpConn := establishVPNConnection()
	defer tcpConn.Close()

	tunIf := setupTUN()
	defer tunIf.Close()

	setupRouting(tunIf.Name)

	pfConf := setupPF(tunIf.Name)
	defer cleanupPF(pfConf)

	handleContextCleanup(ctx, pfConf)

	startForwardingLoop(tunIf, tcpConn)
}

func establishVPNConnection() net.Conn {
	vpnAddr := config.VPNAddress + ":" + strconv.Itoa(config.VPNPort)
	conn, err := tools.EstablishVPNConnection(vpnAddr)
	if err != nil {
		log.Fatalf("failed to establish VPN connection: %v", err)
	}
	log.Printf("VPN connection established to %s", vpnAddr)
	return conn
}

func setupTUN() *tun.TUN {
	tunIf, err := tun.NewTUN()
	if err != nil {
		log.Fatalf("failed to create TUN interface: %v", err)
	}
	log.Printf("TUN interface created: %s", tunIf.Name)
	return tunIf
}

func setupRouting(interfaceName string) {
	if err := route.Setup(interfaceName); err != nil {
		log.Fatalf("route setup failed: %v", err)
	}
	log.Printf("Routing configured for interface: %s", interfaceName)
}

func setupPF(interfaceName string) *pf.Config {
	conf := &pf.Config{
		Interface:  interfaceName,
		ByPassPort: config.VPNPort,
	}
	if err := pf.ApplyRules(conf); err != nil {
		log.Fatalf("PF rule setup failed: %v", err)
	}
	log.Println("PF rules applied")
	return conf
}

func cleanupPF(conf *pf.Config) {
	log.Println("Cleaning up PF rules...")
	if err := pf.RemoveRules(conf); err != nil {
		log.Printf("warning: failed to remove PF rules: %v", err)
	}
}

func handleContextCleanup(ctx context.Context, pfConf *pf.Config) {
	go func() {
		<-ctx.Done()
		cleanupPF(pfConf)
		os.Exit(0)
	}()
}

func startForwardingLoop(tunIf *tun.TUN, conn net.Conn) {
	for {
		if err := tools.ForwardPackets(tunIf, conn); err != nil {
			log.Printf("warning: packet forwarding error: %v", err)
		}
	}
}
