package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/github"
)

const (
	githubRepository = "striker561/Mercury"
	checksumAsset    = "SHA256SUMS"

	updaterAssetDarwin  = "mercury_darwin_universal.zip"
	updaterAssetLinux   = "mercury_linux_amd64"
	updaterAssetWindows = "mercury_windows_amd64.exe"
)

// normalizeVersion strips a leading "v"/"V" so semver comparisons match GitHub
// release tags (v0.2.0) and Wails updater expectations (0.2.0).
func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")
	return v
}

// mercuryAssetMatcher selects the bare-binary / .zip updater artifacts published
// alongside the human-facing installers (.dmg, .deb, NSIS). Wails swaps the
// running binary (or .app bundle on macOS); it cannot apply .dmg/.deb/NSIS.
func mercuryAssetMatcher(req updater.CheckRequest, assets []github.ReleaseAsset) int {
	want := ""
	switch req.Platform {
	case "darwin":
		want = updaterAssetDarwin
	case "linux":
		want = updaterAssetLinux
	case "windows":
		want = updaterAssetWindows
	}
	if want != "" {
		for i, a := range assets {
			if a.Name == want {
				return i
			}
		}
	}
	return github.DefaultAssetMatcher(req, assets)
}

func configureUpdater(app *application.App) error {
	gh, err := github.New(github.Config{
		Repository:    githubRepository,
		ChecksumAsset: checksumAsset,
		AssetMatcher:  mercuryAssetMatcher,
	})
	if err != nil {
		return fmt.Errorf("updater provider: %w", err)
	}
	return app.Updater.Init(updater.Config{
		CurrentVersion: normalizeVersion(Version),
		Providers:      []updater.Provider{gh},
		// Manual "Check for Updates…" in the tray; optional silent startup nudge below.
	})
}

func startSilentUpdateCheck(app *application.App, notify func(title, body string)) {
	go func() {
		time.Sleep(10 * time.Minute)
		rel, err := app.Updater.Check(app.Context())
		if err != nil || rel == nil {
			return
		}
		if notify != nil {
			notify("Mercury", fmt.Sprintf("Version %s is available — use Check for Updates in the tray menu.", rel.Version))
		}
	}()
}
