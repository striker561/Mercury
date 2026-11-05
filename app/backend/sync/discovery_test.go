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

func TestAddrScore(t *testing.T) {
	_, lan, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatal(err)
	}
	locals := []*net.IPNet{lan}

	cases := []struct {
		addr string
		want int
	}{
		{"192.168.1.50:47821", 2},
		{"10.8.0.5:47821", 1},
		{"127.0.0.1:47821", -1},
		{"169.254.1.1:47821", -1},
		{"not-an-address", -1},
	}
	for _, tc := range cases {
		if got := addrScore(tc.addr, locals); got != tc.want {
			t.Errorf("addrScore(%q) = %d, want %d", tc.addr, got, tc.want)
		}
	}
}

func TestEntryAddrsKeepsEveryCandidate(t *testing.T) {
	entry := zeroconf.NewServiceEntry("multi-homed", "_mercury._tcp", "local.")
	entry.Port = 47821
	entry.AddrIPv4 = []net.IP{
		net.ParseIP("10.8.0.5"),
		net.ParseIP("192.168.1.50"),
	}

	addrs := entryAddrs(entry)
	if len(addrs) != 2 {
		t.Fatalf("got %d addresses, want 2: %v", len(addrs), addrs)
	}
	seen := map[string]bool{addrs[0]: true, addrs[1]: true}
	for _, want := range []string{"10.8.0.5:47821", "192.168.1.50:47821"} {
		if !seen[want] {
			t.Fatalf("missing %s in %v", want, addrs)
		}
	}
}
