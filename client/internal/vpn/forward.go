package vpn

import (
	"context"
	"l3vpn-client/internal/config"
	"l3vpn-client/internal/pf"
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

	tcpConn := establishVPNConnection()
	defer tcpConn.Close()

	tunIf := setupTUN()
	defer tunIf.Close()

	pfConf := setupPF(tunIf.Name)
	defer cleanupPF(pfConf)

	handleContextCleanup(ctx, pfConf)

	// TODO: refactor it plz
	go startListeningLoop(tunIf, tcpConn)
	go startForwardingLoop(tunIf, tcpConn)

	log.Print("end forwarding")

	select {}
}

func establishVPNConnection() net.Conn {
	vpnAddr := config.VPNAddress + ":" + strconv.Itoa(config.VPNPort)
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
		util.LogIPv4Packet(packet)

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

func setupPF(interfaceName string) *pf.Config {
	conf := &pf.Config{
		Interface:  interfaceName,
		ByPassIP:   config.VPNAddress,
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
	log.Println("start forwarding")
	buf := make([]byte, 2000)
	for {
		n, err := tunIf.Interface.Read(buf)
		if err != nil {
			log.Printf("warning: tun read fail: %v", err)
			continue
		}
		packet := buf[:n]

		//util.LogIPv4Packet(packet)

		vp := protocol.NewVPNProtocol(packet)
		_, err = conn.Write(vp.Serialize())
		if err != nil {
			log.Printf("warning: packet forwarding error: %v", err)
			continue
		}
	}
}
