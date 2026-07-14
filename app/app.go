package app

import (
	"log"

	"mercury/app/backend/clipboard"
	"mercury/app/backend/storage"
	"mercury/app/services"

	goclipboard "golang.design/x/clipboard"
)

// MercuryApp is the main application struct exposed to the Wails frontend.
type MercuryApp struct {
	syncSvc *services.SyncService
	clipSvc *services.ClipboardService
	db      *storage.DB
}

// NewMercuryApp creates a new MercuryApp instance and opens the settings DB.
func NewMercuryApp() *MercuryApp {
	a := &MercuryApp{}

	p, err := storage.DefaultPath()
	if err != nil {
		log.Printf("[mercury] settings path error: %v", err)
		return a
	}
	db, err := storage.Open(p)
	if err != nil {
		log.Printf("[mercury] settings db error: %v", err)
		return a
	}
	a.db = db

	// Auto-start sync if a passphrase was saved.
	if pass := db.GetPassphrase(); pass != "" {
		a.startSync(pass)
	}

	return a
}

// GetVersion returns the app version string.
func (m *MercuryApp) GetVersion() string {
	return "0.1.0"
}

// SetPassphrase saves the passphrase and starts syncing.
func (m *MercuryApp) SetPassphrase(passphrase string) {
	if m.db != nil {
		m.db.SetPassphrase(passphrase)
	}
	m.startSync(passphrase)
}

// startSync initializes clipboard and begins syncing.
func (m *MercuryApp) startSync(passphrase string) {
	log.Printf("[mercury] starting sync")

	if err := goclipboard.Init(); err != nil {
		log.Printf("[mercury] clipboard init error: %v", err)
		return
	}

	if m.syncSvc != nil {
		m.syncSvc.Stop()
	}
	if m.clipSvc != nil {
		m.clipSvc.Stop()
	}

	m.syncSvc = services.NewSyncService(passphrase)
	if err := m.syncSvc.Start(); err != nil {
		log.Printf("[mercury] sync start error: %v", err)
		return
	}

	m.clipSvc = services.NewClipboardService()
	m.clipSvc.Start(func(c clipboard.Change) {
		if m.db != nil && m.db.IsPaused() {
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
	return m.db != nil && m.db.IsPaused()
}

// TogglePause toggles the sync paused state.
func (m *MercuryApp) TogglePause() bool {
	paused := m.db != nil && m.db.IsPaused()
	paused = !paused
	if m.db != nil {
		m.db.SetPaused(paused)
	}
	if m.clipSvc != nil {
		if paused {
			m.clipSvc.Pause()
		} else {
			m.clipSvc.Resume()
		}
	}
	return paused
}

// GetSavedPassphrase returns the stored passphrase, if any.
func (m *MercuryApp) GetSavedPassphrase() string {
	if m.db == nil {
		return ""
	}
	return m.db.GetPassphrase()
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
	return false
}
