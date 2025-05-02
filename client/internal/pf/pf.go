package pf

import (
	"fmt"
	"l3vpn/internal/util"
	"log"
	"os"
	"strings"
)

const (
	defaultPfConfPath = "/etc/pf.conf"
	ruleBeginComment  = "# vpn-rules BEGIN"
	ruleEndComment    = "# vpn-rules END"
)

type Config struct {
	Interface string // utunx
	Gateway   string // VPN-Gateway IPv4-address
	FilePath  string // pf.conf paths
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

func (c *Config) editConfig() error {
	path := c.getPath()

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := backupFile(path, content); err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	beginIdx, endIdx := c.findRuleBlockIdxs(lines)
	newRules := strings.Split(c.generateRules(), "\n")

	var newLines []string
	if beginIdx != -1 && endIdx != -1 {
		newLines = append(lines[:beginIdx], newRules...)
		newLines = append(newLines, lines[endIdx+1:]...)
	} else {
		newLines = append(lines, newRules...)
	}

	return os.WriteFile(path, []byte(strings.Join(newLines, "\n")), 0644)
}

func (c *Config) cleanConfig() error {
	path := c.getPath()

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	beginIdx, endIdx := c.findRuleBlockIdxs(lines)

	if beginIdx == -1 || endIdx == -1 || beginIdx > endIdx {
		return nil // ничего не удаляем
	}

	newLines := append(lines[:beginIdx], lines[endIdx+1:]...)
	return os.WriteFile(path, []byte(strings.Join(newLines, "\n")), 0644)
}

func (c *Config) findRuleBlockIdxs(lines []string) (int, int) {
	beginIdx, endIdx := -1, -1
	for i, line := range lines {
		if beginIdx == -1 && strings.Contains(line, ruleBeginComment) {
			beginIdx = i
			continue
		}
		if endIdx == -1 && strings.Contains(line, ruleEndComment) {
			endIdx = i
			break
		}
	}
	return beginIdx, endIdx
}

func (c *Config) getPath() string {
	if c.FilePath != "" {
		return c.FilePath
	}
	return defaultPfConfPath
}

func backupFile(path string, content []byte) error {
	return os.WriteFile(path+".bak", content, 0644)
}

func (c *Config) generateRules() string {
	return fmt.Sprintf(`%s
vpn_if = "%s"
vpn_gw = "%s"
pass out route-to ($vpn_if $vpn_gw) from any to any keep state
%s`, ruleBeginComment, c.Interface, c.Gateway, ruleEndComment)
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
