package network

import "l3vpn-client/internal/util"

func RemoveDefaultRoute() error {
	return util.RunCmd("sudo", "route", "delete", "default")
}

func AddDefaultRoute(gw string) error {
	return util.RunCmd("sudo", "route", "add", "default", gw)
}

func AddStaticRoute(host string, gw string) error {
	return util.RunCmd("sudo", "route", "add", "-host", host, gw)
}
