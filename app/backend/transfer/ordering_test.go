package transfer

import (
	"testing"
	"time"
)

// The UI re-fetches the whole dashboard on every progress event, so the
// backend must return offers/transfers in a stable order.  These tests pin
// the arrival-order (oldest first) contract so a future refactor can't
// reintroduce the randomised map-iteration ordering that made the list jump.

func TestPendingOffersSortedByArrival(t *testing.T) {
	m := NewManager(make([]byte, 32))

	first := m.IncomingOffer("b.txt", 10, "192.168.1.10:47821")
	second := m.IncomingOffer("a.txt", 20, "192.168.1.11:47821")
	third := m.IncomingOffer("c.txt", 30, "192.168.1.12:47821")

	// Force a deterministic, non-insertion arrival order.
	m.mu.Lock()
	m.offers[first.ID].CreatedAt = time.Now().Add(-3 * time.Second)
	m.offers[second.ID].CreatedAt = time.Now().Add(-2 * time.Second)
	m.offers[third.ID].CreatedAt = time.Now().Add(-1 * time.Second)
	m.mu.Unlock()

	offers := m.PendingOffers()
	if len(offers) != 3 {
		t.Fatalf("expected 3 offers, got %d", len(offers))
	}
	want := []string{first.ID, second.ID, third.ID}
	for i, id := range want {
		if offers[i].ID != id {
			t.Fatalf("offers[%d].ID = %q, want %q", i, offers[i].ID, id)
		}
	}
}

func TestPendingOffersTieBreakByID(t *testing.T) {
	m := NewManager(make([]byte, 32))
	base := time.Now()

	m.mu.Lock()
	m.offers["z-id"] = &Offer{ID: "z-id", FileName: "z", CreatedAt: base}
	m.offers["a-id"] = &Offer{ID: "a-id", FileName: "a", CreatedAt: base}
	m.mu.Unlock()

	offers := m.PendingOffers()
	if len(offers) != 2 {
		t.Fatalf("expected 2 offers, got %d", len(offers))
	}
	if offers[0].ID != "a-id" || offers[1].ID != "z-id" {
		t.Fatalf("expected ID tie-break, got %q then %q", offers[0].ID, offers[1].ID)
	}
}

func TestAllProgressSortedByArrival(t *testing.T) {
	m := NewManager(make([]byte, 32))

	m.mu.Lock()
	m.transfers["t-a"] = &Progress{ID: "t-a", FileName: "a", Status: StatusSending, CreatedAt: time.Now().Add(-3 * time.Second)}
	m.transfers["t-b"] = &Progress{ID: "t-b", FileName: "b", Status: StatusReceiving, CreatedAt: time.Now().Add(-2 * time.Second)}
	m.transfers["t-c"] = &Progress{ID: "t-c", FileName: "c", Status: StatusSending, CreatedAt: time.Now().Add(-1 * time.Second)}
	// Done transfers must be omitted entirely.
	m.transfers["t-d"] = &Progress{ID: "t-d", FileName: "d", Status: StatusDone, CreatedAt: time.Now().Add(-4 * time.Second)}
	m.mu.Unlock()

	progs := m.AllProgress()
	if len(progs) != 3 {
		t.Fatalf("expected 3 active transfers, got %d", len(progs))
	}
	want := []string{"t-a", "t-b", "t-c"}
	for i, id := range want {
		if progs[i].ID != id {
			t.Fatalf("progs[%d].ID = %q, want %q", i, progs[i].ID, id)
		}
	}
}
