package network

import (
	"fmt"
	"strconv"

	"l3vpn-client/internal/util"
)

func EnableInfe(infeName string, local string, gw string, mtu int) error {
	if err := util.RunCmd("sudo", "ifconfig", infeName, local, gw, "mtu", strconv.Itoa(mtu), "up"); err != nil {
		return fmt.Errorf("failed to enable %s interface  : %w", infeName, err)
	}
	return nil
}

func DisableInfe(infeName string) error {
	if err := util.RunCmd("sudo", "ifconfig", infeName, "down"); err != nil {
		return fmt.Errorf("failed to disable %s interface  : %w", infeName, err)
	}
	return nil
}
