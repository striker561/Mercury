package transfer

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCleanupStaleParts(t *testing.T) {
	dir := t.TempDir()
	oldPart := filepath.Join(dir, "file.bin.part-old")
	freshPart := filepath.Join(dir, "file.bin.part-fresh")
	finished := filepath.Join(dir, "file.bin")

	for _, path := range []string{oldPart, freshPart, finished} {
		if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	oldTime := time.Now().Add(-stalePartAge - time.Hour)
	if err := os.Chtimes(oldPart, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	cleanupStaleParts(dir, time.Now())

	if _, err := os.Stat(oldPart); !os.IsNotExist(err) {
		t.Fatal("stale .part file was not removed")
	}
	for _, path := range []string{freshPart, finished} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("%s was removed: %v", path, err)
		}
	}
}
