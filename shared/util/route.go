package util

func RemoveDefaultRoute() error {
	return RunCmd("sudo", "route", "delete", "default")
}

func AddDefaultRoute(gw string) error {
	return RunCmd("sudo", "route", "add", "default", gw)
}

func AddStaticRoute(host string, gw string) error {
	return RunCmd("sudo", "route", "add", "-host", host, gw)
}
