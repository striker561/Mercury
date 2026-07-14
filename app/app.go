package app

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"mercury/app/backend/clipboard"
	"mercury/app/backend/crypto"
	"mercury/app/backend/fileinfo"
	"mercury/app/backend/storage"
	"mercury/app/services"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"
	goclipboard "golang.design/x/clipboard"
)

// MercuryApp is the main application struct exposed to the Wails frontend.
type MercuryApp struct {
	syncSvc    *services.SyncService
	clipSvc    *services.ClipboardService
	db         *storage.DB
	transSvc   *services.TransferService
	showWindow func()
	hideWindow func()
	emitChange func()
	notifySvc  *notifications.NotificationService
	autostart  *application.AutostartManager
	syncAt     time.Time // last clipboard sync activity
	gnomeTray  bool      // GNOME detected — tray may need AppIndicator
}

// SetShowWindow registers a callback to show the settings window.
func (m *MercuryApp) SetShowWindow(fn func()) {
	m.showWindow = fn
}

// SetHideWindow registers a callback to hide the settings window.
func (m *MercuryApp) SetHideWindow(fn func()) {
	m.hideWindow = fn
}

// SetEmitChange registers a callback invoked when dashboard state changes.
func (m *MercuryApp) SetEmitChange(fn func()) {
	m.emitChange = fn
}

// HideWindow hides the settings window (tray app stays running).
func (m *MercuryApp) HideWindow() {
	if m.hideWindow != nil {
		m.hideWindow()
	}
}

// SetNotifier stores the notification service for OS-level alerts.
func (m *MercuryApp) SetNotifier(ns *notifications.NotificationService) {
	m.notifySvc = ns
}

// SetGNOMETrayTip marks that the app is running under GNOME.
func (m *MercuryApp) SetGNOMETrayTip(enabled bool) {
	m.gnomeTray = enabled
}

// SetAutostartManager stores the OS autostart manager.
func (m *MercuryApp) SetAutostartManager(am *application.AutostartManager) {
	m.autostart = am
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
	m.notifyChange()
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
	//
	// IMPORTANT: transport listener delivers RAW CIPHERTEXT for all message
	// types.  Clipboard is decrypted by sync's event loop; file chunks must
	// be decrypted here before feeding into the transfer manager.
	key := services.DeriveKey(passphrase)
	m.transSvc = services.NewTransferService(key)
	m.syncSvc.SetOnMessage(func(msgType byte, payload []byte) {
		dec, err := crypto.Decrypt(payload, key)
		if err != nil {
			log.Printf("[mercury] decrypt chunk: %v", err)
			return
		}
		m.transSvc.OnWireMessage(msgType, dec)
	})
	m.syncSvc.SetOnFileOffer(func(offerID, fileName string, fileSize int64, peerAddr string) {
		// Use the sender's offer ID so file_accept maps back correctly.
		m.transSvc.IncomingOfferWithID(offerID, fileName, fileSize, peerAddr)
		// Auto-accept if the setting is on.
		if m.db != nil && m.db.GetDefaulted(storage.KeyAutoAccept) == "true" {
			saveDir := resolvePath(m.GetReceivedFolder())
			m.transSvc.AcceptOffer(offerID, saveDir)
		}
		m.notifyChange()
		if m.showWindow != nil {
			m.showWindow()
		}
		// Fire an OS notification so the user knows a file arrived.
		if m.notifySvc != nil {
			go m.notifySvc.SendNotification(notifications.NotificationOptions{
				Title: "Mercury",
				Body:  fmt.Sprintf("Incoming file: %s (%d MB)", fileName, fileSize/1024/1024+1),
			})
		}
	})
	// When a peer accepts our file offer, look up the file path and start sending.
	m.syncSvc.SetOnFileAccept(func(offerID string) {
		if fp := m.transSvc.AcceptNotification(offerID); fp != "" {
			log.Printf("[mercury] offer %s accepted, sending %s", offerID, fp)
			// Take first peer — in practice there's only one sender per offer.
			peers := m.syncSvc.GetPeers()
			if len(peers) > 0 {
				m.transSvc.SendFileForOffer(offerID, peers[0]["addr"], fp)
			}
		}
	})

	if err := m.syncSvc.Start(); err != nil {
		log.Printf("[mercury] sync start error: %v", err)
		return
	}

	m.syncClipboardWatch()
}

func (m *MercuryApp) shouldWatchClipboard() bool {
	if m.syncSvc == nil || m.GetSavedPassphrase() == "" || m.IsPaused() {
		return false
	}
	return m.GetPeerCount() > 0
}

// syncClipboardWatch starts clipboard monitoring only when peers are connected.
// No passphrase, no peers, or paused — Mercury rests. No polling.
func (m *MercuryApp) syncClipboardWatch() {
	if !m.shouldWatchClipboard() {
		if m.clipSvc != nil {
			m.clipSvc.Stop()
			m.clipSvc = nil
		}
		return
	}
	if m.clipSvc == nil {
		m.startClipboardWatcher()
	}
}

func (m *MercuryApp) startClipboardWatcher() {
	m.clipSvc = services.NewClipboardService()
	m.clipSvc.Start(func(c clipboard.Change) {
		if m.db != nil && m.db.IsPaused() {
			return
		}
		switch c.Type {
		case clipboard.ChangeText:
			if fi := fileinfo.Detect(c.Text); fi != nil {
				log.Printf("[mercury] detected file: %s (%d bytes, %s)", fi.Name, fi.Size, fi.Category)
				id := m.transSvc.NewOfferID()
				m.transSvc.StoreOutgoing(id, fi.Path)
				m.syncSvc.BroadcastFileOffer(id, fi.Name, fi.Size)
				m.MarkSyncActivity()
				return
			}
			m.syncSvc.BroadcastText(c.Text)
			m.MarkSyncActivity()
		case clipboard.ChangeImage:
			m.syncSvc.BroadcastImage(c.Image)
			m.MarkSyncActivity()
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
	m.notifyChange()
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
	if key == storage.KeyAutostart && m.autostart != nil {
		if value == "true" {
			if err := m.autostart.Enable(); err != nil {
				log.Printf("[mercury] autostart enable: %v", err)
			}
		} else {
			if err := m.autostart.Disable(); err != nil {
				log.Printf("[mercury] autostart disable: %v", err)
			}
		}
	}
}

// GetReceivedFolder returns the path where received files are stored.
func (m *MercuryApp) GetReceivedFolder() string {
	p := m.GetSetting(storage.KeyReceivedFolder)
	if p == "" {
		return "~/Downloads/Mercury/"
	}
	return p
}

// SetReceivedFolder saves the received file path.
func (m *MercuryApp) SetReceivedFolder(path string) {
	if m.db != nil {
		m.db.Set(storage.KeyReceivedFolder, path)
	}
}

// PickReceivedFolder opens a native folder picker dialog and returns the
// selected path.  Returns empty string if the user cancels.
func (m *MercuryApp) PickReceivedFolder() string {
	// Try zenity first (Linux), then osascript (macOS).
	for _, args := range [][]string{
		{"zenity", "--file-selection", "--directory", "--title=Select received files folder"},
		{"osascript", "-e", `choose folder with prompt "Select received files folder"`},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		out, err := cmd.Output()
		if err != nil {
			continue
		}
		p := strings.TrimSpace(string(out))
		if p != "" && !strings.HasPrefix(p, "FAIL") {
			// osascript returns "alias Macintosh HD:Users:..." — strip alias prefix
			if strings.HasPrefix(p, "alias ") {
				p = strings.TrimPrefix(p, "alias ")
			}
			// macOS returns colon-delimited paths; convert to POSIX
			p = strings.ReplaceAll(p, ":", "/")
			if !strings.HasPrefix(p, "/") {
				p = "/" + p
			}
			// Store immediately so next GetReceivedFolder returns it.
			m.SetReceivedFolder(p)
			return p
		}
	}
	return ""
}

// resolvePath expands a leading ~/ to the current user's home directory.
// This is needed because Go's os package does not interpret tilde.
func resolvePath(p string) string {
	if p == "" {
		p = "~/Mercury"
	}
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Printf("[mercury] resolve home dir: %v", err)
			return p
		}
		return home + p[1:]
	}
	return p
}

// MarkSyncActivity records that we just sent clipboard data to peers.
func (m *MercuryApp) MarkSyncActivity() {
	m.syncAt = time.Now()
}

// HasRecentSync returns true if clipboard was synced within the last 2 seconds.
func (m *MercuryApp) HasRecentSync() bool {
	return time.Since(m.syncAt) < 2*time.Second
}

// trayActive returns true when the tray icon should show the active state.
func (m *MercuryApp) trayActive() bool {
	if m.HasRecentSync() {
		return true
	}
	for _, p := range m.GetTransferProgress() {
		if p.Status == "sending" || p.Status == "receiving" {
			return true
		}
	}
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
	saveDir := resolvePath(m.GetReceivedFolder())
	tid, err := m.transSvc.AcceptOffer(offerID, saveDir)
	if err != nil {
		log.Printf("[mercury] accept offer: %v", err)
		return ""
	}
	log.Printf("[mercury] accepted offer %s, transfer %s", offerID, tid)
	// Tell the sender we accepted so they start streaming the file.
	m.syncSvc.BroadcastFileAccept(offerID)
	m.notifyChange()
	return tid
}

// RejectFileOffer rejects an incoming file offer.
func (m *MercuryApp) RejectFileOffer(offerID string) {
	if m.transSvc != nil {
		m.transSvc.RejectOffer(offerID)
	}
	m.notifyChange()
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

// CancelTransfer cancels a running file transfer.
func (m *MercuryApp) CancelTransfer(tid string) {
	if m.transSvc != nil {
		m.transSvc.CancelTransfer(tid)
	}
	m.notifyChange()
}
