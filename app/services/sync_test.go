package services

import "testing"

func TestPeerAddrFromRemoteUsesListenerPort(t *testing.T) {
	if got, want := peerAddrFromRemote("192.168.1.9:55555"), "192.168.1.9:47821"; got != want {
		t.Fatalf("peerAddrFromRemote = %q, want %q", got, want)
	}
}
