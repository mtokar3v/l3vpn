package connection

import (
	"l3vpn-server/internal/nat"
	"net"
	"strconv"
)

func PublicSocket(c net.Conn) nat.Socket {
	addrSrt := c.RemoteAddr().String()
	orgAddr, portStr, _ := net.SplitHostPort(addrSrt)
	orgPort, _ := strconv.Atoi(portStr)

	return nat.Socket{
		IPAddr: orgAddr,
		Port:   uint16(orgPort),
	}
}

func PrivateSocket(c net.Conn) nat.Socket {
	addrSrt := c.LocalAddr().String()
	_, portStr, _ := net.SplitHostPort(addrSrt)
	orgPort, _ := strconv.Atoi(portStr)

	return nat.Socket{
		IPAddr: "127.0.0.1",
		Port:   uint16(orgPort),
	}
}
