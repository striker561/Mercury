package sync

import (
	"net"
	"testing"

	"github.com/grandcat/zeroconf"
)

func TestPeerIDFromEntry(t *testing.T) {
	entry := zeroconf.NewServiceEntry("DESKTOP-ABC", "_mercury._tcp", "local.")
	entry.Text = []string{"id=deadbeef1234"}
	if got := peerIDFromEntry(entry); got != "deadbeef1234" {
		t.Fatalf("peerIDFromEntry = %q, want deadbeef1234", got)
	}
}

func TestPeerIDFromEntryMissing(t *testing.T) {
	// Legacy peer: no TXT metadata -> no machine ID.
	entry := zeroconf.NewServiceEntry("legacy-peer", "_mercury._tcp", "local.")
	if got := peerIDFromEntry(entry); got != "" {
		t.Fatalf("expected empty ID for legacy peer, got %q", got)
	}
}

func TestResolveEntryPrefersMachineID(t *testing.T) {
	entry := zeroconf.NewServiceEntry("mercury-laptop", "_mercury._tcp", "local.")
	entry.Text = []string{"id=uuid-123"}
	entry.AddrIPv4 = []net.IP{net.ParseIP("192.168.1.50")}
	entry.Port = 47821

	p := resolveEntry(entry)
	if p == nil {
		t.Fatal("resolveEntry returned nil")
	}
	if p.ID != "uuid-123" {
		t.Fatalf("ID = %q, want machine id uuid-123", p.ID)
	}
	if p.Hostname != "mercury-laptop" {
		t.Fatalf("Hostname = %q, want announced hostname mercury-laptop", p.Hostname)
	}
}

func TestResolveEntryLegacyFallback(t *testing.T) {
	entry := zeroconf.NewServiceEntry("legacy-host", "_mercury._tcp", "local.")
	entry.AddrIPv4 = []net.IP{net.ParseIP("192.168.1.51")}
	entry.Port = 47821

	p := resolveEntry(entry)
	if p == nil {
		t.Fatal("resolveEntry returned nil")
	}
	if p.ID != "legacy-host" || p.Hostname != "legacy-host" {
		t.Fatalf("legacy fallback wrong: ID=%q Hostname=%q", p.ID, p.Hostname)
	}
}
