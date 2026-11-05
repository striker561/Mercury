package transfer

import "testing"

func TestHasActiveTransfers(t *testing.T) {
	m := NewManager(nil)

	if m.HasActiveTransfers() {
		t.Fatal("empty manager reports an active transfer")
	}

	m.mu.Lock()
	m.transfers["done"] = &Progress{Status: StatusDone}
	m.transfers["failed"] = &Progress{Status: StatusFailed}
	m.mu.Unlock()
	if m.HasActiveTransfers() {
		t.Fatal("terminal transfers must not block a network resync")
	}

	m.mu.Lock()
	m.transfers["accepting"] = &Progress{Status: StatusAccepting}
	m.mu.Unlock()
	if !m.HasActiveTransfers() {
		t.Fatal("accepting transfer must block a network resync")
	}
}

func TestOfferPeerReturnsSender(t *testing.T) {
	m := NewManager(nil)
	m.IncomingOfferWithID("0123456789abcdef", "file.bin", 1, "10.0.0.5:47821")

	if got := m.OfferPeer("0123456789abcdef"); got != "10.0.0.5:47821" {
		t.Fatalf("OfferPeer = %q, want the sender address", got)
	}
}
