package connection

import (
	"l3vpn-server/internal/nat"
	"net"
	"strconv"
)

func RemoteSocket(c net.Conn) nat.Socket {
	addrSrt := c.RemoteAddr().String()
	orgAddr, portStr, _ := net.SplitHostPort(addrSrt)
	orgPort, _ := strconv.Atoi(portStr)

	return nat.Socket{
		IPAddr: orgAddr,
		Port:   uint16(orgPort),
	}
}

func LocalSocket(c net.Conn) nat.Socket {
	addrSrt := c.LocalAddr().String()
	orgAddr, portStr, _ := net.SplitHostPort(addrSrt)
	orgPort, _ := strconv.Atoi(portStr)

	return nat.Socket{
		IPAddr: orgAddr,
		Port:   uint16(orgPort),
	}
}
