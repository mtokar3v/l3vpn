package route

import (
	"fmt"

	"l3vpn/internal/util"
)

type Config struct {
	Interface string
	LocalIP   string
	Gateway   string
}

func (c *Config) Setup() error {
	if err := util.RunCmd("sudo", "ifconfig", c.Interface, c.LocalIP, c.Gateway, "up"); err != nil {
		return fmt.Errorf("failed to add default route: %w", err)
	}
	return nil
}
