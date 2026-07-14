package storage

import (
	"os"
	"testing"
)

func TestOpenClose(t *testing.T) {
	db, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()
}

func TestGetSet(t *testing.T) {
	db := openMem(t)
	defer db.Close()

	// Get non-existent key returns empty.
	val, err := db.Get("nonexistent")
	if err != nil {
		t.Fatalf("Get nonexistent failed: %v", err)
	}
	if val != "" {
		t.Fatalf("expected '', got %q", val)
	}

	// Set and get.
	if err := db.Set("passphrase", "hunter2"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	val, err = db.Get("passphrase")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "hunter2" {
		t.Fatalf("expected 'hunter2', got %q", val)
	}
}

func TestUpdate(t *testing.T) {
	db := openMem(t)
	defer db.Close()

	db.Set("sync_enabled", "true")
	db.Set("sync_enabled", "false")

	val, _ := db.Get("sync_enabled")
	if val != "false" {
		t.Fatalf("expected 'false', got %q", val)
	}
}

func TestMultipleKeys(t *testing.T) {
	db := openMem(t)
	defer db.Close()

	db.Set("key1", "val1")
	db.Set("key2", "val2")
	db.Set("key3", "val3")

	v1, _ := db.Get("key1")
	v2, _ := db.Get("key2")
	v3, _ := db.Get("key3")

	if v1 != "val1" || v2 != "val2" || v3 != "val3" {
		t.Fatal("multiple keys mismatch")
	}
}

func TestFilePersistence(t *testing.T) {
	path := tempFile(t)
	defer os.Remove(path)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	db.Set("passphrase", "test123")
	db.Close()

	// Reopen and verify data persisted.
	db2, err := Open(path)
	if err != nil {
		t.Fatalf("Reopen failed: %v", err)
	}
	defer db2.Close()

	val, _ := db2.Get("passphrase")
	if val != "test123" {
		t.Fatalf("expected 'test123', got %q", val)
	}
}

func openMem(t *testing.T) *DB {
	t.Helper()
	db, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	return db
}

func tempFile(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "mercury-test-*.db")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	f.Close()
	return f.Name()
}
