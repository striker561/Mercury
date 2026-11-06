package transfer

import "testing"

func TestPruneFailedTransfersKeepsOthers(t *testing.T) {
	m := NewManager(nil)
	m.mu.Lock()
	m.transfers["failed"] = &Progress{ID: "failed", FileName: "file.bin", Status: StatusFailed}
	m.transfers["done"] = &Progress{ID: "done", FileName: "file.bin", Status: StatusDone}
	m.transfers["active"] = &Progress{ID: "active", FileName: "file.bin", Status: StatusSending}
	m.transfers["other"] = &Progress{ID: "other", FileName: "other.bin", Status: StatusFailed}
	m.pruneFailedTransfersLocked("file.bin")
	m.mu.Unlock()

	if m.Progress("failed") != nil {
		t.Fatal("previous failure was not pruned")
	}
	for _, id := range []string{"done", "active", "other"} {
		if m.Progress(id) == nil {
			t.Fatalf("%s transfer was pruned", id)
		}
	}
}
