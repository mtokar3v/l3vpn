package pf

import (
	"fmt"
	"l3vpn-client/internal/util"
	"log"
)

type Config struct {
	Interface string // utunX
	Gateway   string // VPN-Gateway IP
	FilePath  string // pf.conf path override
}

func (c *Config) ApplyRules() error {
	if err := c.editConfig(); err != nil {
		return fmt.Errorf("failed to edit pf.conf: %w", err)
	}
	if err := c.validate(); err != nil {
		return fmt.Errorf("pf.conf validation failed: %w", err)
	}
	if err := c.reload(); err != nil {
		return fmt.Errorf("pfctl reload failed: %w", err)
	}
	if err := c.enable(); err != nil {
		log.Printf("warning: failed to enable pf: %v", err)
	}
	return nil
}

func (c *Config) RemoveRules() error {
	if err := c.cleanConfig(); err != nil {
		return fmt.Errorf("failed to clean pf.conf: %w", err)
	}
	if err := c.validate(); err != nil {
		return fmt.Errorf("pf.conf validation failed: %w", err)
	}
	if err := c.reload(); err != nil {
		return fmt.Errorf("pfctl reload failed: %w", err)
	}
	return nil
}

func (c *Config) validate() error {
	return util.RunCmd("sudo", "pfctl", "-n", "-f", c.getPath())
}

func (c *Config) reload() error {
	return util.RunCmd("sudo", "pfctl", "-f", c.getPath())
}

func (c *Config) enable() error {
	return util.RunCmd("sudo", "pfctl", "-e")
}
