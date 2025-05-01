package pf

import (
	"fmt"
	"l3vpn/internal/util"
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
		fmt.Printf("warning: failed to enable pf: %v\n", err)
	}
	return nil
}

func (c *Config) editConfig() error {
	path := c.FilePath
	if path == "" {
		path = defaultPfConfPath
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	backupPath := path + ".bak"
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	pfText := string(content)
	ruleBlock := c.generateRules()

	beginIdx := -1
	endIdx := -1
	lines := strings.Split(pfText, "\n")
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

	if beginIdx != -1 {
		ruleLines := strings.Split(ruleBlock, "\n")

		start := beginIdx
		end := endIdx + 1

		newLines := append(lines[:start], ruleLines...)
		newLines = append(newLines, lines[end:]...)
		pfText = strings.Join(newLines, "\n")
	} else {
		pfText += "\n" + ruleBlock
	}

	return os.WriteFile(path, []byte(pfText), 0644)
}

func (c *Config) generateRules() string {
	return fmt.Sprintf(`%s
vpn_if = "%s"
vpn_gw = "%s"
pass out route-to ($vpn_if $vpn_gw) from any to any keep state
%s`, ruleBeginComment, c.Interface, c.Gateway, ruleEndComment)
}

func (c *Config) validate() error {
	path := c.FilePath
	if path == "" {
		path = defaultPfConfPath
	}
	return util.RunCmd("sudo", "pfctl", "-n", "-f", path)
}

func (c *Config) reload() error {
	path := c.FilePath
	if path == "" {
		path = defaultPfConfPath
	}
	return util.RunCmd("sudo", "pfctl", "-f", path)
}

func (c *Config) enable() error {
	return util.RunCmd("sudo", "pfctl", "-e")
}
