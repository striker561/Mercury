package app

import "testing"

func TestIsVPNInterface(t *testing.T) {
	if !isVPNInterface("utun4") {
		t.Fatal("utun4 should match")
	}
	if !isVPNInterface("WindscribeWireguard") {
		t.Fatal("Windscribe should match")
	}
	if isVPNInterface("en0") {
		t.Fatal("en0 should not match")
	}
}
