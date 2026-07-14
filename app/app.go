package app

import (
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"mercury/app/backend/clipboard"
	"mercury/app/backend/storage"
	"mercury/app/services"

	goclipboard "golang.design/x/clipboard"
)

// MercuryApp is the main application struct exposed to the Wails frontend.
type MercuryApp struct {
	syncSvc     *services.SyncService
	clipSvc     *services.ClipboardService
	db          *storage.DB
	transSvc    *services.TransferService
	showWindow  func()
}

// SetShowWindow registers a callback to show the settings window.
func (m *MercuryApp) SetShowWindow(fn func()) {
	m.showWindow = fn
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
	return Version
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

	// Wire the shared TCP listener: sync handles clipboard, transfer handles
	// file chunks.  They don't know about each other — OnMessage is the glue.
	key := services.DeriveKey(passphrase)
	m.transSvc = services.NewTransferService(key)
	m.syncSvc.SetOnMessage(func(msgType byte, payload []byte) {
		m.transSvc.ChunkChan() <- payload
	})
	m.syncSvc.SetOnFileOffer(func(offerID, fileName string, fileSize int64, peerAddr string) {
		// Use the sender's offer ID so file_accept maps back correctly.
		m.transSvc.IncomingOfferWithID(offerID, fileName, fileSize, peerAddr)
		// Auto-accept if the setting is on.
		if m.db != nil && m.db.GetDefaulted(storage.KeyAutoAccept) == "true" {
			saveDir := m.GetReceivedFolder()
			if saveDir == "" {
				saveDir = "~/Mercury"
			}
			m.transSvc.AcceptOffer(offerID, saveDir)
		}
	})
	// When a peer accepts our file offer, look up the file path and start sending.
	m.syncSvc.SetOnFileAccept(func(offerID string) {
		if fp := m.transSvc.AcceptNotification(offerID); fp != "" {
			log.Printf("[mercury] offer %s accepted, sending %s", offerID, fp)
			// Take first peer — in practice there's only one sender per offer.
			peers := m.syncSvc.GetPeers()
			if len(peers) > 0 {
				m.transSvc.SendFile(peers[0]["addr"], fp)
			}
		}
	})

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
			filePath := c.Text
			// macOS Finder copies file:// URIs; Linux file managers often do too.
			if strings.HasPrefix(filePath, "file://") {
				if u, err := url.Parse(filePath); err == nil {
					filePath = u.Path
				}
			}
			// If the text is a valid file path, offer it as a file transfer.
			if fi, err := os.Stat(filePath); err == nil && !fi.IsDir() {
				name := filepath.Base(filePath)
				log.Printf("[mercury] detected file path: %s (%d bytes)", name, fi.Size())
				// Generate an offer ID and broadcast it so the receiver uses the
				// same ID — when they file_accept, we can look up the file path.
				id := m.transSvc.NewOfferID()
				m.transSvc.StoreOutgoing(id, filePath)
				m.syncSvc.BroadcastFileOffer(id, name, fi.Size())
				return
			}
			log.Printf("[mercury] clipboard text: %d chars", len(c.Text))
			m.syncSvc.BroadcastText(c.Text)
		case clipboard.ChangeImage:
			log.Printf("[mercury] clipboard image: %d bytes", len(c.Image))
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

// GetAllSettings returns all known settings plus the app version
// in a single IPC round-trip.
//
// Why batch? Every Wails IPC call has overhead (JSON serialisation,
// Go->JS marshalling, event loop tick). Loading 30 settings individually
// would be 30 round-trips. One call with a map[string]string is always
// faster — the overhead of sending a few extra bytes is negligible
// compared to 29 extra bridge crossings.
//
// Similarly, storage.DB.All() runs one SQL query, not one per key.
func (m *MercuryApp) GetAllSettings() map[string]string {
	out := map[string]string{"version": Version}
	if m.db != nil {
		s := m.db.All() // single SQL query, fills in defaults
		for k, v := range s {
			out[k] = v
		}
	} else {
		for _, k := range storage.AllKeys() {
			out[k] = storage.Default(k)
		}
	}
	return out
}

// GetSetting returns any setting by key, falling back to its default.
func (m *MercuryApp) GetSetting(key string) string {
	if m.db == nil {
		return storage.Default(key)
	}
	return m.db.GetDefaulted(key)
}

// SetSetting saves any setting by key.
func (m *MercuryApp) SetSetting(key, value string) {
	if m.db != nil {
		m.db.Set(key, value)
	}
	// Special handling for settings that need runtime action.
	if key == storage.KeyAllowFiles && value == "false" && m.clipSvc != nil {
		m.clipSvc.Pause()
	}
}

// GetReceivedFolder returns the path where received files are stored.
func (m *MercuryApp) GetReceivedFolder() string {
	return m.GetSetting(storage.KeyReceivedFolder)
}

// DetectGNOME returns true if running under GNOME desktop.
func (m *MercuryApp) DetectGNOME() bool {
	return false
}

// ─── File Transfer IPC ──────────────────────────────────────────────

// GetPendingFileOffers returns all file offers that haven't been acted on.
func (m *MercuryApp) GetPendingFileOffers() []services.FileOffer {
	if m.transSvc == nil {
		return nil
	}
	return m.transSvc.PendingOffers()
}

// AcceptFileOffer accepts an incoming file offer and starts receiving.
func (m *MercuryApp) AcceptFileOffer(offerID string) string {
	if m.transSvc == nil || m.syncSvc == nil {
		return ""
	}
	saveDir := m.GetReceivedFolder()
	if saveDir == "" {
		saveDir = "~/Mercury"
	}
	tid, err := m.transSvc.AcceptOffer(offerID, saveDir)
	if err != nil {
		log.Printf("[mercury] accept offer: %v", err)
		return ""
	}
	log.Printf("[mercury] accepted offer %s, transfer %s", offerID, tid)
	// Tell the sender we accepted so they start streaming the file.
	m.syncSvc.BroadcastFileAccept(offerID)
	return tid
}

// RejectFileOffer rejects an incoming file offer.
func (m *MercuryApp) RejectFileOffer(offerID string) {
	if m.transSvc != nil {
		m.transSvc.RejectOffer(offerID)
	}
}

// SendFile sends a file to a peer.  Returns the transfer ID.
func (m *MercuryApp) SendFile(peerAddr, filePath string) string {
	if m.transSvc == nil {
		return ""
	}
	tid, err := m.transSvc.SendFile(peerAddr, filePath)
	if err != nil {
		log.Printf("[mercury] send file: %v", err)
		return ""
	}
	return tid
}

// GetTransferProgress returns progress for all active transfers.
func (m *MercuryApp) GetTransferProgress() []services.FileProgress {
	if m.transSvc == nil {
		return nil
	}
	return m.transSvc.AllProgress()
}
