package sync

import (
	"testing"
	"time"
)

func TestPeerMapAddOrUpdate(t *testing.T) {
	pm := NewPeerMap()
	defer pm.Stop()

	pm.AddOrUpdate("peer1", "192.168.1.10:47821")
	pm.AddOrUpdate("peer2", "192.168.1.11:47821")

	if pm.Len() != 2 {
		t.Fatalf("expected 2 peers, got %d", pm.Len())
	}

	peers := pm.GetPeers()
	ids := make(map[string]bool)
	for _, p := range peers {
		ids[p.ID] = true
	}
	if !ids["peer1"] || !ids["peer2"] {
		t.Fatal("missing expected peers")
	}
}

func TestPeerMapUpdateRefreshesLastSeen(t *testing.T) {
	pm := NewPeerMap()
	defer pm.Stop()

	pm.AddOrUpdate("peer1", "192.168.1.10:47821")
	time.Sleep(10 * time.Millisecond)
	pm.AddOrUpdate("peer1", "192.168.1.10:47822") // different port

	peers := pm.GetPeers()
	if len(peers) != 1 {
		t.Fatalf("expected 1 peer after update, got %d", len(peers))
	}
	if peers[0].Addr != "192.168.1.10:47822" {
		t.Fatalf("expected updated addr, got %s", peers[0].Addr)
	}
}

func TestPeerMapRemove(t *testing.T) {
	pm := NewPeerMap()
	defer pm.Stop()

	pm.AddOrUpdate("peer1", "192.168.1.10:47821")
	pm.Remove("peer1")

	if pm.Len() != 0 {
		t.Fatal("expected 0 peers after removal")
	}
}

func TestPeerMapRecordFailureEvictsAfterThree(t *testing.T) {
	pm := NewPeerMap()
	defer pm.Stop()

	pm.AddOrUpdate("peer1", "192.168.1.10:47821")

	// First two failures should not evict.
	if pm.RecordFailure("peer1") {
		t.Fatal("first failure should not evict")
	}
	if pm.RecordFailure("peer1") {
		t.Fatal("second failure should not evict")
	}

	// Third failure should evict.
	if !pm.RecordFailure("peer1") {
		t.Fatal("third failure should evict")
	}

	if pm.Len() != 0 {
		t.Fatal("expected 0 peers after 3 failures")
	}
}

func TestPeerMapResetFailures(t *testing.T) {
	pm := NewPeerMap()
	defer pm.Stop()

	pm.AddOrUpdate("peer1", "192.168.1.10:47821")

	pm.RecordFailure("peer1")
	pm.RecordFailure("peer1")
	pm.ResetFailures("peer1")

	// After reset, should need 3 more failures to evict.
	if pm.RecordFailure("peer1") {
		t.Fatal("first failure after reset should not evict")
	}
	if pm.RecordFailure("peer1") {
		t.Fatal("second failure after reset should not evict")
	}
	if !pm.RecordFailure("peer1") {
		t.Fatal("third failure after reset should evict")
	}
}

func TestPeerMapAddOrUpdateResetsFailures(t *testing.T) {
	pm := NewPeerMap()
	defer pm.Stop()

	pm.AddOrUpdate("peer1", "192.168.1.10:47821")
	pm.RecordFailure("peer1")
	pm.RecordFailure("peer1")

	// Re-announcement via mDNS should reset failures.
	pm.AddOrUpdate("peer1", "192.168.1.10:47821")

	if pm.RecordFailure("peer1") {
		t.Fatal("first failure after re-announce should not evict")
	}
	if pm.RecordFailure("peer1") {
		t.Fatal("second failure after re-announce should not evict")
	}
	if !pm.RecordFailure("peer1") {
		t.Fatal("third failure after re-announce should evict")
	}
}

func TestPeerMapGetPeersSnapshot(t *testing.T) {
	pm := NewPeerMap()
	defer pm.Stop()

	pm.AddOrUpdate("peer1", "192.168.1.10:47821")

	peers := pm.GetPeers()
	peers[0].ID = "modified"

	// Original should be unchanged (snapshot copy).
	if pm.Len() != 1 {
		t.Fatal("GetPeers should return a copy, not a reference")
	}
}

func TestPeerMapGetPeersSorted(t *testing.T) {
	pm := NewPeerMap()
	defer pm.Stop()

	// Insert out of alphabetical order on purpose — the map's iteration order
	// is randomised, so GetPeers must sort to keep the UI list stable.
	pm.AddOrUpdate("zeta", "192.168.1.3:47821")
	pm.AddOrUpdate("alpha", "192.168.1.1:47821")
	pm.AddOrUpdate("mike", "192.168.1.2:47821")

	peers := pm.GetPeers()
	if len(peers) != 3 {
		t.Fatalf("expected 3 peers, got %d", len(peers))
	}
	want := []string{"alpha", "mike", "zeta"}
	for i, id := range want {
		if peers[i].ID != id {
			t.Fatalf("peers[%d].ID = %q, want %q (order must be stable)", i, peers[i].ID, id)
		}
	}
}

func TestPeerMapAddOrUpdatePeerDedupByMachineID(t *testing.T) {
	pm := NewPeerMap()
	defer pm.Stop()

	// Same physical machine, two OSes: same machine ID, same IP, different
	// hostnames.  Dedup by machine ID must collapse them into one entry.
	pm.AddOrUpdatePeer(Peer{ID: "mach-1", Hostname: "DESKTOP-ABC", Addr: "192.168.1.10:47821"})
	pm.AddOrUpdatePeer(Peer{ID: "mach-1", Hostname: "mercury-laptop", Addr: "192.168.1.10:47821"})

	if pm.Len() != 1 {
		t.Fatalf("expected 1 peer after machine-ID dedup, got %d", pm.Len())
	}
	peers := pm.GetPeers()
	if len(peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(peers))
	}
	// The hostname should reflect the latest announcement (the OS that just
	// booted), while the ID stays the stable machine ID.
	if peers[0].ID != "mach-1" {
		t.Fatalf("ID = %q, want stable machine ID mach-1", peers[0].ID)
	}
	if peers[0].Hostname != "mercury-laptop" {
		t.Fatalf("Hostname = %q, want latest announcement mercury-laptop", peers[0].Hostname)
	}
}

func TestPeerMapDistinctMachinesNotMerged(t *testing.T) {
	pm := NewPeerMap()
	defer pm.Stop()

	// Two genuinely different machines must stay separate even if they have
	// the same hostname (each has its own machine ID).
	pm.AddOrUpdatePeer(Peer{ID: "mach-a", Hostname: "office-pc", Addr: "192.168.1.20:47821"})
	pm.AddOrUpdatePeer(Peer{ID: "mach-b", Hostname: "office-pc", Addr: "192.168.1.21:47821"})

	if pm.Len() != 2 {
		t.Fatalf("expected 2 distinct peers, got %d", pm.Len())
	}
}
