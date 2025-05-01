package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/songgao/water"
)

func main() {
	vpnGw := "10.0.0.3"

	tun, err := createTunInteface()
	if err != nil {
		panic(err)
	}

	log.Printf("Interface Name: %s\n", tun.Name())

	if err := setUpTunRouting(tun, vpnGw); err != nil {
		panic(err)
	}

	if err := editPf(tun.Name(), vpnGw); err != nil {
		panic(err)
	}

	if err := refreshPf(); err != nil {
		panic(err)
	}

	if err := listenTun(tun); err != nil {
		panic(err)
	}
}

func createTunInteface() (ifce *water.Interface, err error) {
	return water.New(water.Config{
		DeviceType: water.TUN,
	})
}

func setUpTunRouting(tun *water.Interface, vpnGw string) error {
	kernelIpv4 := "10.0.0.1"

	err := exec.Command("sudo", "ifconfig", tun.Name(), kernelIpv4, vpnGw, "up").Run()
	if err != nil {
		return err
	}

	exec.Command("route", "add", "default", vpnGw).Run()
	if err != nil {
		return err
	}

	return nil
}

func listenTun(tun *water.Interface) error {

	packet := make([]byte, 2000)

	for {
		n, err := tun.Read(packet)
		if err != nil {
			return err
		}
		log.Printf("Packet Received: % x\n", packet[:n])
	}
}

func editPf(vpnIf string, vpnGw string) error {
	cfgPath := "/etc/pf.conf"
	cmt := "# my-vpn-rules"
	settings := fmt.Sprintf(`
%s
vpn_if = "%s"
vpn_gw = "%s"
pass out route-to ($vpn_if $vpn_gw) from any to any keep state`, cmt, vpnIf, vpnGw)
	settings = strings.Trim(settings, "\n")

	cnt, err := os.ReadFile(cfgPath)
	if err != nil {
		return err
	}

	pfText := string(cnt)
	idx := strings.Index(pfText, cmt)
	if idx == -1 {
		pfText += settings
	} else {
		lines := strings.Split(pfText, "\n")

		for i, line := range lines {
			if strings.Contains(line, cmt) {
				idx = i
				break
			}
		}

		pfLines := strings.Split(settings, "\n")

		tmp := append(lines[:idx], pfLines...)
		lines = append(tmp, lines[idx+len(pfLines):]...)

		pfText = strings.Join(lines, "\n")
	}

	if err := os.WriteFile(cfgPath, []byte(pfText), 0644); err != nil {
		return err
	}

	return nil
}

func refreshPf() error {
	cfgPath := "/etc/pf.conf"

	// Проверка синтаксиса
	if err := exec.Command("sudo", "pfctl", "-n", "-f", cfgPath).Run(); err != nil {
		return fmt.Errorf("syntax error in pf.conf: %w", err)
	}

	// Применение правил
	if err := exec.Command("sudo", "pfctl", "-f", cfgPath).Run(); err != nil {
		return fmt.Errorf("failed to load pf.conf: %w", err)
	}

	// Включение pf
	if err := exec.Command("sudo", "pfctl", "-e").Run(); err != nil {
		fmt.Printf("warning during enabling pf: %w", err)
	}

	return nil
}
