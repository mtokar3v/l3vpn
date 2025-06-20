package pf

import (
	"fmt"
	"l3vpn-client/internal/config"
	"os"
	"strings"
)

const (
	ruleBeginComment = "# vpn-rules BEGIN"
	ruleEndComment   = "# vpn-rules END"
)

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
		return nil // nothing to clean
	}

	newLines := append(lines[:beginIdx], lines[endIdx+1:]...)
	return os.WriteFile(path, []byte(strings.Join(newLines, "\n")), 0644)
}

func (c *Config) generateRules() string {
	return fmt.Sprintf(`%s
vpn_if = "%s"
vpn_gw = "%s"
by_pass_ip = "%s"
by_pass_port = "%d"

pass out quick on egress proto tcp from any to $by_pass_ip port $by_pass_port keep state
pass out route-to ($vpn_if $vpn_gw) from any to any keep state
%s`, ruleBeginComment, c.Interface, config.Gateway, c.ByPassIP, c.ByPassPort, ruleEndComment)
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
