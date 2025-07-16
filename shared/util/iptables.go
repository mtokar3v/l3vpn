package util

func FlushNat() error {
	return RunCmd("sudo", "iptables", "-t", "nat", "-F")
}

func Snat(o string, ip string) error {
	return RunCmd("sudo", "iptables", "-t", "nat", "-A", "POSTROUTING", "-o", o, "-j", "SNAT", "--to-source", ip)
}

func AcceptForwarding() error {
	return RunCmd("sudo", "iptables", "-P", "FORWARD", "ACCEPT")
}
