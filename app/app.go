package app

import (
	"log"

	"mercury/app/backend/clipboard"
	"mercury/app/services"

	goclipboard "golang.design/x/clipboard"
)

// MercuryApp is the main application struct exposed to the Wails frontend.
// Business logic is delegated to services/. This stays thin.
type MercuryApp struct {
	passphrase string
	paused     bool
	syncSvc    *services.SyncService
	clipSvc    *services.ClipboardService
}

// NewMercuryApp creates a new MercuryApp instance.
func NewMercuryApp() *MercuryApp {
	return &MercuryApp{}
}

// GetVersion returns the app version string.
func (m *MercuryApp) GetVersion() string {
	return "0.1.0"
}

// SetPassphrase stores the sync passphrase and starts sync + clipboard watching.
func (m *MercuryApp) SetPassphrase(passphrase string) {
	m.passphrase = passphrase
	log.Printf("[mercury] passphrase updated")

	// Initialize the clipboard library (safe to call multiple times)
	if err := goclipboard.Init(); err != nil {
		log.Printf("[mercury] clipboard init error: %v", err)
		return
	}

	// Stop previous services if restarting
	if m.syncSvc != nil {
		m.syncSvc.Stop()
	}
	if m.clipSvc != nil {
		m.clipSvc.Stop()
	}

	// Create and start sync service
	m.syncSvc = services.NewSyncService(passphrase)
	if err := m.syncSvc.Start(); err != nil {
		log.Printf("[mercury] sync start error: %v", err)
		return
	}

	// Create and start clipboard service
	// When clipboard changes, broadcast to all peers
	m.clipSvc = services.NewClipboardService()
	m.clipSvc.Start(func(c clipboard.Change) {
		if m.paused {
			return
		}
		switch c.Type {
		case clipboard.ChangeText:
			m.syncSvc.BroadcastText(c.Text)
		case clipboard.ChangeImage:
			m.syncSvc.BroadcastImage(c.Image)
		}
	})
}

// IsPaused returns whether sync is paused.
func (m *MercuryApp) IsPaused() bool {
	return m.paused
}

// TogglePause toggles the sync paused state.
func (m *MercuryApp) TogglePause() bool {
	m.paused = !m.paused
	if m.clipSvc != nil {
		if m.paused {
			m.clipSvc.Pause()
		} else {
			m.clipSvc.Resume()
		}
	}
	return m.paused
}

// GetPeerCount returns the number of connected peers.
func (m *MercuryApp) GetPeerCount() int {
	if m.syncSvc == nil {
		return 0
	}
	return m.syncSvc.PeerCount()
}

// GetPeers returns the list of connected peers.
func (m *MercuryApp) GetPeers() []map[string]string {
	if m.syncSvc == nil {
		return nil
	}
	return m.syncSvc.GetPeers()
}

// GetReceivedFolder returns the path where received files are stored.
func (m *MercuryApp) GetReceivedFolder() string {
	return "~/Mercury/"
}

// DetectGNOME returns true if running under GNOME desktop.
func (m *MercuryApp) DetectGNOME() bool {
	return false // real detection in main.go via env vars
}
