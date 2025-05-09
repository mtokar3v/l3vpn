package route

import (
	"fmt"

	"l3vpn-client/internal/util"
)

type Config struct {
	Interface string
	LocalIP   string
	Gateway   string
}

func Setup(c *Config) error {
	if err := util.RunCmd("sudo", "ifconfig", c.Interface, c.LocalIP, c.Gateway, "up"); err != nil {
		return fmt.Errorf("failed to add default route: %w", err)
	}
	return nil
}
