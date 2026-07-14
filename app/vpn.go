package app

import (
	"net"
	"strings"
)

// vpnActive reports whether a VPN-style network interface is up.
func vpnActive() bool {
	ifaces, err := net.Interfaces()
	if err != nil {
		return false
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if isVPNInterface(iface.Name) {
			return true
		}
	}
	return false
}

func isVPNInterface(name string) bool {
	n := strings.ToLower(name)
	for _, p := range []string{
		"tun", "tap", "wg", "utun", "ppp", "ipsec", "vpn",
		"windscribe", "nordlynx", "mullvad", "proton", "tailscale", "zerotier",
	} {
		if strings.HasPrefix(n, p) || strings.Contains(n, p) {
			return true
		}
	}
	return false
}
