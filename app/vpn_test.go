package app

import "testing"

func TestIsVPNInterface(t *testing.T) {
	// macOS system tunnels — not VPN
	for _, name := range []string{"utun0", "utun4", "en0", "bridge0", "awdl0"} {
		if isVPNInterface(name) {
			t.Fatalf("%s should not be VPN", name)
		}
	}

	// Known clients
	if !isVPNInterface("WindscribeWireguard") {
		t.Fatal("Windscribe should match")
	}
	if !isVPNInterface("wg0") {
		t.Fatal("wg0 should match")
	}
	if !isVPNInterface("tun0") {
		t.Fatal("tun0 should match")
	}
	if isVPNInterface("utun0") {
		t.Fatal("utun0 must not match")
	}
}

func TestVpnActiveNotFalsePositiveOnTypicalMacNames(t *testing.T) {
	names := []string{"utun0", "utun1", "utun2", "utun3", "en0", "lo0"}
	for _, name := range names {
		if isVPNInterface(name) {
			t.Fatalf("false positive for %s", name)
		}
	}
	if vpnActive() {
		t.Fatal("vpnActive() should be false on a machine with only system utun interfaces")
	}
}
