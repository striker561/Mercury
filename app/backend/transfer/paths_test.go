package transfer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSafeName(t *testing.T) {
	cases := map[string]string{
		"photo.png":        "photo.png",
		"../../etc/passwd": "passwd",
		`..\..\secret.txt`: "secret.txt",
		"":                 "download",
		".":                "download",
		"..":               "download",
	}
	for in, want := range cases {
		if got := safeName(in); got != want {
			t.Errorf("safeName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestUniqueDestination(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(first, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	got := uniqueDestination(first)
	if got != filepath.Join(dir, "file (1).txt") {
		t.Fatalf("uniqueDestination = %q", got)
	}
}
