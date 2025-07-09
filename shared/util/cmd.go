package util

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func RunCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Running: %s %s\n", name, strings.Join(args, " "))

	return cmd.Run()
}
