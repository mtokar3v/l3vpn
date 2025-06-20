package pf

import (
	"fmt"
	"l3vpn-client/internal/util"
	"log"
)

type Config struct {
	Interface  string // utunX
	FilePath   string // pf.conf path override
	ByPassIP   string
	ByPassPort int
}

func ApplyRules(c *Config) error {
	if err := c.editConfig(); err != nil {
		return fmt.Errorf("failed to edit pf.conf: %w", err)
	}
	if err := validate(c); err != nil {
		return fmt.Errorf("pf.conf validation failed: %w", err)
	}
	if err := reload(c); err != nil {
		return fmt.Errorf("pfctl reload failed: %w", err)
	}
	if err := enable(c); err != nil {
		log.Printf("warning: failed to enable pf: %v", err)
	}
	return nil
}

func RemoveRules(c *Config) error {
	if err := c.cleanConfig(); err != nil {
		return fmt.Errorf("failed to clean pf.conf: %w", err)
	}
	if err := validate(c); err != nil {
		return fmt.Errorf("pf.conf validation failed: %w", err)
	}
	if err := reload(c); err != nil {
		return fmt.Errorf("pfctl reload failed: %w", err)
	}
	return nil
}

func validate(c *Config) error {
	return util.RunCmd("sudo", "pfctl", "-n", "-f", c.getPath())
}

func reload(c *Config) error {
	return util.RunCmd("sudo", "pfctl", "-f", c.getPath())
}

func enable(c *Config) error {
	return util.RunCmd("sudo", "pfctl", "-e")
}
