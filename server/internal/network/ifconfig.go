package network

import (
	"fmt"
	"l3vpn-server/internal/config"
	"l3vpn-server/internal/util"
)

func Enable(infeName string) error {
	if err := util.RunCmd("sudo", "ifconfig", infeName, config.LocalIP, config.Gateway, "mtu", "1300", "up"); err != nil {
		return fmt.Errorf("failed to enable %s interface  : %w", infeName, err)
	}
	return nil
}

func Disable(infeName string) error {
	if err := util.RunCmd("sudo", "ifconfig", infeName, config.LocalIP, config.Gateway, "down"); err != nil {
		return fmt.Errorf("failed to disable %s interface  : %w", infeName, err)
	}
	return nil
}
