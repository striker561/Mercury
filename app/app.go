package app

import (
	"log"
)

// MercuryApp is the main application struct exposed to the Wails frontend.
type MercuryApp struct {
	passphrase string
	paused     bool
}

// NewMercuryApp creates a new MercuryApp instance.
func NewMercuryApp() *MercuryApp {
	return &MercuryApp{}
}

// GetVersion returns the app version string.
func (m *MercuryApp) GetVersion() string {
	return "0.1.0"
}

// SetPassphrase stores the sync passphrase.
func (m *MercuryApp) SetPassphrase(passphrase string) {
	m.passphrase = passphrase
	log.Printf("[mercury] passphrase updated")
}

// IsPaused returns whether sync is paused.
func (m *MercuryApp) IsPaused() bool {
	return m.paused
}

// TogglePause toggles the sync paused state.
func (m *MercuryApp) TogglePause() bool {
	m.paused = !m.paused
	return m.paused
}

// GetPeerCount returns the current number of connected peers.
// Phase 1 returns 0 — will be wired to sync manager in Phase 2.
func (m *MercuryApp) GetPeerCount() int {
	return 0
}

// GetPeers returns the list of connected peers.
// Phase 1 returns empty — will be wired in Phase 2.
func (m *MercuryApp) GetPeers() []map[string]string {
	return []map[string]string{}
}

// GetReceivedFolder returns the path where received files are stored.
func (m *MercuryApp) GetReceivedFolder() string {
	return "~/Mercury/"
}

// DetectGNOME returns true if running under GNOME desktop.
func (m *MercuryApp) DetectGNOME() bool {
	return false // simplified; real detection in main.go via env vars
}
