package transfer

import (
	"fmt"
	"testing"
	"time"
)

func TestCancelTransferIsIdempotent(t *testing.T) {
	m := NewManager(nil)
	m.cancel["tid"] = make(chan struct{})

	m.CancelTransfer("tid")
	m.CancelTransfer("tid")
}

func TestPruneTerminalTransfersKeepsActive(t *testing.T) {
	m := NewManager(nil)
	now := time.Now()

	m.mu.Lock()
	for i := 0; i < maxTerminalTransfers+10; i++ {
		id := fmt.Sprintf("done-%d", i)
		m.transfers[id] = &Progress{ID: id, Status: StatusDone, CreatedAt: now.Add(time.Duration(i) * time.Second)}
	}
	m.transfers["active"] = &Progress{ID: "active", Status: StatusSending, CreatedAt: now}
	m.pruneTerminalLocked()
	m.mu.Unlock()

	terminal := 0
	for _, p := range m.transfers {
		if p.Status == StatusDone {
			terminal++
		}
	}
	if terminal > maxTerminalTransfers {
		t.Fatalf("kept %d terminal transfers, want at most %d", terminal, maxTerminalTransfers)
	}
	if m.transfers["active"] == nil {
		t.Fatal("active transfer was pruned")
	}
}

func TestPruneOutgoingExpiresOldOffers(t *testing.T) {
	m := NewManager(nil)

	m.mu.Lock()
	m.outgoing["old"] = outgoingOffer{path: "/tmp/old", at: time.Now().Add(-outgoingOfferTTL - time.Minute)}
	m.outgoing["fresh"] = outgoingOffer{path: "/tmp/fresh", at: time.Now()}
	m.pruneOutgoingLocked()
	m.mu.Unlock()

	if _, ok := m.outgoing["old"]; ok {
		t.Fatal("expired outgoing offer was kept")
	}
	if _, ok := m.outgoing["fresh"]; !ok {
		t.Fatal("fresh outgoing offer was pruned")
	}
}

func TestPruneOutgoingCapsSize(t *testing.T) {
	m := NewManager(nil)

	m.mu.Lock()
	for i := 0; i < maxOutgoingOffers+10; i++ {
		id := fmt.Sprintf("offer-%d", i)
		m.outgoing[id] = outgoingOffer{path: "/tmp/file", at: time.Now()}
	}
	m.pruneOutgoingLocked()
	m.mu.Unlock()

	if len(m.outgoing) > maxOutgoingOffers {
		t.Fatalf("kept %d outgoing offers, want at most %d", len(m.outgoing), maxOutgoingOffers)
	}
}
