package route

import (
	"fmt"

	"l3vpn-client/internal/config"
	"l3vpn-client/internal/util"
)

func Setup(tunName string) error {
	if err := util.RunCmd("sudo", "ifconfig", tunName, config.LocalIP, config.Gateway, "up"); err != nil {
		return fmt.Errorf("failed to add default route: %w", err)
	}
	return nil
}
