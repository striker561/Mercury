package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mercury/app/backend/storage"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/github"
)

const (
	githubRepository = "striker561/Mercury"
	checksumAsset    = "SHA256SUMS"
	updateCheckEvery = 24 * time.Hour

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
		HTTPClient:    &http.Client{Timeout: 10 * time.Minute, Transport: retryTransport{}},
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

func startSilentUpdateCheck(app *application.App, mercuryApp *MercuryApp) {
	if mercuryApp == nil || !mercuryApp.updateCheckDue() {
		return
	}

	go func() {
		time.Sleep(10 * time.Minute)
		if !mercuryApp.updateCheckDue() {
			return
		}

		ctx, cancel := context.WithTimeout(app.Context(), 20*time.Second)
		defer cancel()

		rel, err := app.Updater.Check(ctx)
		mercuryApp.markUpdateChecked()
		if err != nil {
			log.Printf("[mercury] update check: %v", err)
			return
		}
		if rel == nil {
			return
		}
		mercuryApp.notify("Mercury", fmt.Sprintf("Version %s is available — use Check for Updates in the tray menu.", rel.Version))
	}()
}

func updateCheckDue(last string, now time.Time) bool {
	if last == "" {
		return true
	}
	sec, err := strconv.ParseInt(last, 10, 64)
	if err != nil || sec <= 0 {
		return true
	}
	return now.Sub(time.Unix(sec, 0)) >= updateCheckEvery
}

func (m *MercuryApp) updateCheckDue() bool {
	if m.db == nil {
		return true
	}
	return updateCheckDue(m.db.GetDefaulted(storage.KeyLastUpdateCheck), time.Now())
}

func (m *MercuryApp) markUpdateChecked() {
	if m.db != nil {
		m.db.Set(storage.KeyLastUpdateCheck, strconv.FormatInt(time.Now().Unix(), 10))
	}
}
