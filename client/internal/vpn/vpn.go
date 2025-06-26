package vpn

import (
	"context"
	"l3vpn-client/internal/config"
	"l3vpn-client/internal/network"
	"l3vpn-client/internal/protocol"
	"l3vpn-client/internal/tun"
	"l3vpn-client/internal/util"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

func Start(ctx context.Context) {
	log.Println("start forwarding")

	tunIf := setupTUN()
	tcpConn := establishVPNConnection()
	setRoutes()

	log.Printf("tun ifce: %s", tunIf.Name)

	handleContextCleanup(ctx, tunIf, tcpConn)

	go startListeningLoop(tunIf, tcpConn)
	go startForwardingLoop(tunIf, tcpConn)

	select {}
}

func establishVPNConnection() net.Conn {
	vpnAddr := config.VPNServerAddress + ":" + strconv.Itoa(config.VPNServerPort)
	log.Printf("try to connect to %s", vpnAddr)
	conn, err := net.Dial("tcp4", vpnAddr)
	if err != nil {
		log.Fatalf("failed to establish VPN connection: %v", err)
	}
	log.Printf("VPN connection established to %s", vpnAddr)

	time.Sleep(5 * time.Second)

	return conn
}

func startListeningLoop(tunIf *tun.TUN, conn net.Conn) {
	log.Println("start listening")
	buf := make([]byte, 2000)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("listening error %v", err)
			continue
		}

		packet := buf[:n]
		util.LogIPv4Packet("[INBOUND]", packet)

		_, err = tunIf.Interface.Write(packet)
		if err != nil {
			log.Printf("warning: tun write fail: %v", err)
			continue
		}
	}
}

func setupTUN() *tun.TUN {
	tunIf, err := tun.NewTUN()
	if err != nil {
		log.Fatalf("failed to create TUN interface: %v", err)
	}
	log.Printf("TUN interface created: %s", tunIf.Name)

	return tunIf
}

func setRoutes() {
	err := network.RemoveDefaultRoute()
	if err != nil {
		log.Fatalf("failed to delete default route: %v", err)
	}

	err = network.AddDefaultRoute(config.TUNGateway)
	if err != nil {
		log.Fatalf("failed to add default route: %v", err)
	}

	err = network.AddStaticRoute(config.VPNServerAddress, config.DefaultGateway)
	if err != nil {
		log.Fatalf("failed to add static route to vpn: %v", err)
	}
}

func handleContextCleanup(ctx context.Context, tunIf *tun.TUN, conn net.Conn) {
	go func() {
		<-ctx.Done()
		tunIf.Close()
		conn.Close()
		network.RemoveDefaultRoute()
		network.AddDefaultRoute(config.DefaultGateway)
		os.Exit(0)
	}()
}

func startForwardingLoop(tunIf *tun.TUN, conn net.Conn) {
	log.Println("start forwarding")
	buf := make([]byte, 2000)
	for {
		n, err := tunIf.Interface.Read(buf)
		if err != nil {
			log.Printf("warning: tun read fail: %v", err)
			continue
		}
		packet := buf[:n]

		util.LogIPv4Packet("[OUTBOUND]", packet)

		vp := protocol.NewVPNProtocol(packet)
		_, err = conn.Write(vp.Serialize())
		if err != nil {
			log.Printf("warning: packet forwarding error: %v", err)
			continue
		}
	}
}
