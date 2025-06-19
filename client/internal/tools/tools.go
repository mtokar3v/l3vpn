package tools

import (
	"fmt"
	"l3vpn-client/internal/protocol"
	"l3vpn-client/internal/tun"
	"l3vpn-client/internal/util"
	"net"
)

const packetBufferSize = 2000 // in bytes

func ForwardPackets(t *tun.TUN, conn net.Conn) error {
	buf := make([]byte, packetBufferSize)
	for {
		n, err := t.Interface.Read(buf)
		if err != nil {
			return err
		}
		packet := buf[:n]

		util.LogIPv4Packet(packet)

		vp := protocol.NewVPNProtocol(packet)
		_, err = conn.Write(vp.Serialize())
		if err != nil {
			return fmt.Errorf("failed to write into connection: %v", err)
		}
	}
}

func EstablishVPNConnection(addr string) (net.Conn, error) {
	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
