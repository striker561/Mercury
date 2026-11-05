package app

import (
	"strconv"
	"testing"
	"time"

	"github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/github"
)

func TestNormalizeVersion(t *testing.T) {
	tests := map[string]string{
		"v0.2.0":   "0.2.0",
		"V1.0.0":   "1.0.0",
		"0.2.0":    "0.2.0",
		"  v1.2.3": "1.2.3",
	}
	for in, want := range tests {
		if got := normalizeVersion(in); got != want {
			t.Errorf("normalizeVersion(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestUpdateCheckDue(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	stamp := func(d time.Duration) string {
		return strconv.FormatInt(now.Add(d).Unix(), 10)
	}

	tests := []struct {
		name string
		last string
		want bool
	}{
		{name: "never checked", last: "", want: true},
		{name: "invalid timestamp", last: "not-a-number", want: true},
		{name: "checked recently", last: stamp(-23 * time.Hour), want: false},
		{name: "checked a day ago", last: stamp(-24 * time.Hour), want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := updateCheckDue(tt.last, now); got != tt.want {
				t.Fatalf("updateCheckDue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMercuryAssetMatcherPrefersUpdaterArtifacts(t *testing.T) {
	assets := []github.ReleaseAsset{
		{Name: "mercury-macos-universal.dmg"},
		{Name: updaterAssetDarwin},
		{Name: "mercury_linux_amd64.deb"},
		{Name: updaterAssetLinux},
		{Name: "mercury-installer.exe"},
		{Name: updaterAssetWindows},
		{Name: checksumAsset},
	}

	if got := mercuryAssetMatcher(updater.CheckRequest{Platform: "darwin", Arch: "arm64"}, assets); got != 1 {
		t.Fatalf("darwin: index %d, want 1", got)
	}
	if got := mercuryAssetMatcher(updater.CheckRequest{Platform: "linux", Arch: "amd64"}, assets); got != 3 {
		t.Fatalf("linux: index %d, want 3", got)
	}
	if got := mercuryAssetMatcher(updater.CheckRequest{Platform: "windows", Arch: "amd64"}, assets); got != 5 {
		t.Fatalf("windows: index %d, want 5", got)
	}
}
