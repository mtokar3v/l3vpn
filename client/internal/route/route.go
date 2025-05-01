package route

import (
	"fmt"
	"os/exec"
)

type Config struct {
	InterfaceName string
	LocalIP       string
	Gateway       string
}

func (c *Config) Setup() error {
	if err := exec.Command("sudo", "ifconfig", c.InterfaceName, c.LocalIP, c.Gateway, "up").Run(); err != nil {
		return fmt.Errorf("failed to add default route: %w", err)
	}
	return nil
}
