package vpn

import (
	"context"
	"l3vpn/client/config"
	"l3vpn/shared/tun"
	"l3vpn/shared/util"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

func Start(ctx context.Context) {
	log.Println("start forwarding")
	tun := setupTUN()
	tcpConn := establishVPNConnection()
	if err := setupNat(tun); err != nil {
		log.Fatalf("failed to set up NAT: %v", err)
	}
	setRoutes()
	log.Printf("tun ifce: %s", tun.Name)
	handleContextCleanup(ctx, tun, tcpConn)
	go startListeningLoop(tun, tcpConn)
	go startForwardingLoop(tun, tcpConn)
	select {}
}

func establishVPNConnection() *net.TCPConn {
	vpnAddr := config.VPNServerAddress + ":" + strconv.Itoa(config.VPNServerPort)
	log.Printf("try to connect to %s", vpnAddr)
	conn, err := net.Dial("tcp4", vpnAddr)
	if err != nil {
		log.Fatalf("failed to establish VPN connection: %v", err)
	}

	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		log.Fatal("not a TCP connection")
	}

	log.Printf("VPN connection established to %s", vpnAddr)

	time.Sleep(5 * time.Second)

	return tcpConn
}

func startListeningLoop(tunIf *tun.Tun, conn *net.TCPConn) {
	log.Println("start listening")
	for {
		packet, err := util.ReadPacket(conn)
		if err != nil {
			log.Printf("listening error %v", err)
			continue
		}
		util.LogIPv4Packet("[INBOUND]", packet)
		_, err = tunIf.Infe.Write(packet)
		if err != nil {
			log.Printf("warning: tun write fail: %v", err)
			continue
		}
	}
}

func setupTUN() *tun.Tun {
	tunIf, err := tun.NewTun()
	if err != nil {
		log.Fatalf("failed to create TUN interface: %v", err)
	}
	err = tunIf.Up()
	if err != nil {
		log.Fatalf("failed to up tun interface: %v", err)
	}
	log.Printf("TUN interface created: %s", tunIf.Name)
	return tunIf
}

func setupNat(t *tun.Tun) error {
	if err := util.FlushNat(); err != nil {
		return err
	}
	// All incoming packets with tun ip pass to tun infe
	// NAT return origin daddr and kernel could handle packet normally
	if err := util.Snat(t.Name, config.TUNGateway); err != nil {
		return err
	}
	// tunx -> enx
	if err := util.AcceptForwarding(); err != nil {
		return err
	}
	return nil
}

func setRoutes() {
	err := util.RemoveDefaultRoute()
	if err != nil {
		log.Fatalf("failed to delete default route: %v", err)
	}

	err = util.AddDefaultRoute(config.TUNGateway)
	if err != nil {
		log.Fatalf("failed to add default route: %v", err)
	}

	err = util.AddStaticRoute(config.VPNServerAddress, config.DefaultGateway)
	if err != nil {
		log.Fatalf("failed to add static route to vpn: %v", err)
	}
}

func handleContextCleanup(ctx context.Context, tunIf *tun.Tun, conn net.Conn) {
	go func() {
		<-ctx.Done()
		tunIf.Close()
		conn.Close()
		util.RemoveDefaultRoute()
		util.AddDefaultRoute(config.DefaultGateway)
		os.Exit(0)
	}()
}

func startForwardingLoop(tunIf *tun.Tun, conn *net.TCPConn) {
	log.Println("start forwarding")
	buf := make([]byte, 2000)
	for {
		n, err := tunIf.Infe.Read(buf)
		if err != nil {
			log.Printf("warning: tun read fail: %v", err)
			continue
		}
		packet := buf[:n]

		// x := gopacket.NewPacket(packet, layers.LayerTypeIPv4, gopacket.Default)

		// ipLayer := x.Layer(layers.LayerTypeIPv4)
		// if ipLayer == nil {
		// 	log.Printf("Not an IPv4 packet")
		// 	return
		// }

		// ip, ok := ipLayer.(*layers.IPv4)
		// if !ok {
		// 	log.Printf("Failed to cast to IPv4 layer")
		// 	return
		// }

		// if ip.DstIP.String() != "146.190.62.39" {
		// 	// ignore
		// 	continue
		// }

		util.LogIPv4Packet("[OUTBOUND]", packet)
		_, err = util.WritePacket(conn, packet)
		if err != nil {
			log.Printf("warning: packet forwarding error: %v", err)
			continue
		}
	}
}
