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
