package app

import (
	"net"
	"strings"
)

// vpnActive reports whether a known VPN client interface appears to be up.
// Conservative on purpose: macOS always has utun* for system services — never
// treat those as VPN.
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

	// Named VPN clients (interface name often includes the product).
	for _, needle := range []string{
		"windscribe", "nordlynx", "nordvpn", "mullvad", "proton",
		"tailscale", "zerotier", "wireguard", "openvpn", "surfshark",
		"expressvpn", "cyberghost", "pia-", "privateinternetaccess",
	} {
		if strings.Contains(n, needle) {
			return true
		}
	}

	// WireGuard on Linux: wg0, wg1, …
	if strings.HasPrefix(n, "wg") && len(n) >= 3 && n[2] >= '0' && n[2] <= '9' {
		return true
	}

	// OpenVPN on Linux: tun0, tap0 (not utun — that prefix is macOS system).
	if len(n) >= 4 && (strings.HasPrefix(n, "tun") || strings.HasPrefix(n, "tap")) {
		suffix := n[3:]
		if suffix == "" || (len(suffix) == 1 && suffix[0] >= '0' && suffix[0] <= '9') {
			return !strings.HasPrefix(n, "utun")
		}
	}

	return false
}
