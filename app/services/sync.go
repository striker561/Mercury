package services

import (
	"context"
	"encoding/json"
	"log"

	"mercury/app/backend/crypto"
	"mercury/app/backend/sync"

	goclipboard "golang.design/x/clipboard"
)

// wirePayload is the JSON structure sent over the network.
type wirePayload struct {
	Type    string `json:"type"`    // "text", "image", "file_offer", "file_accept"
	Content string `json:"content"` // text content (for text type)
	Image   []byte `json:"image"`   // PNG bytes (for image type)

	// File offer / accept fields.
	OfferID  string `json:"offer_id,omitempty"`
	FileName string `json:"file_name,omitempty"`
	FileSize int64  `json:"file_size,omitempty"`
}

// OnFileOfferCallback is called when a file offer arrives from a peer.
type OnFileOfferCallback func(offerID, fileName string, fileSize int64, peerAddr string)

// OnFileAcceptCallback is called when a remote peer accepts our file offer.
type OnFileAcceptCallback func(offerID string)

// SyncService manages LAN peer discovery, clipboard broadcast, and receive.
type SyncService struct {
	manager       *sync.Manager
	onFileOffer   OnFileOfferCallback
	onFileAccept  OnFileAcceptCallback
	onResync      func()
	resyncAllowed func() bool
	watchCancel   context.CancelFunc
}

// SetOnMessage registers a callback for non-clipboard messages (file chunks)
// received on the shared TCP listener.
func (s *SyncService) SetOnMessage(handler func(byte, []byte)) {
	s.manager.OnMessage = handler
}

// NewSyncService creates a sync service with the given passphrase and this
// machine's stable device ID (announced via mDNS TXT for dual-boot dedup).
func NewSyncService(passphrase, deviceID string) *SyncService {
	return &SyncService{
		manager: sync.NewManager(passphrase, deviceID),
	}
}

// DeriveKey derives the encryption key from the passphrase.
func DeriveKey(passphrase string) []byte {
	return crypto.DeriveKey(passphrase)
}

// SetOnFileOffer registers a callback for incoming file offers.
func (s *SyncService) SetOnFileOffer(cb OnFileOfferCallback) {
	s.onFileOffer = cb
}

// SetOnFileAccept registers a callback when a peer accepts our file offer.
func (s *SyncService) SetOnFileAccept(cb OnFileAcceptCallback) {
	s.onFileAccept = cb
}

// BroadcastFileAccept sends an accept notification for an offer back to peers.
func (s *SyncService) BroadcastFileAccept(offerID string) {
	p := wirePayload{Type: "file_accept", OfferID: offerID}
	data, _ := json.Marshal(p)
	s.manager.Broadcast(data)
}

// SetOnResync registers a callback fired after an automatic resync, so the
// app layer can refresh the UI.
func (s *SyncService) SetOnResync(cb func()) {
	s.onResync = cb
}

// SetResyncAllowed registers a predicate that must return true for automatic
// resyncs to proceed.  The app uses it to skip resync while a file transfer
// is active — the shared TCP listener would be restarted mid-stream.
func (s *SyncService) SetResyncAllowed(fn func() bool) {
	s.resyncAllowed = fn
}

// Resync restarts discovery immediately.
func (s *SyncService) Resync() {
	if err := s.manager.Resync(); err != nil {
		log.Printf("[sync] resync: %v", err)
	}
}

// startNetworkWatch begins automatic discovery retries: when the network
// interfaces change or no peers are found for a while, discovery is
// restarted (this is what the manual toggle used to do).  Resyncs are
// skipped while resyncAllowed() returns false.
func (s *SyncService) startNetworkWatch() {
	ctx, cancel := context.WithCancel(context.Background())
	s.watchCancel = cancel
	w := sync.NewNetworkWatcher()
	go w.Start(ctx, func() int { return s.manager.PeerCount() }, func() {
		if s.resyncAllowed != nil && !s.resyncAllowed() {
			return
		}
		if err := s.manager.Resync(); err != nil {
			log.Printf("[sync] resync: %v", err)
			return
		}
		if s.onResync != nil {
			s.onResync()
		}
	})
}

// Start begins mDNS discovery, TCP listening, and automatic network-watch
// retries (resync when interfaces change or no peers are found for a while).
func (s *SyncService) Start() error {
	s.manager.SetOnReceive(func(payload []byte) {
		var p wirePayload
		if err := json.Unmarshal(payload, &p); err != nil {
			log.Printf("[sync] unmarshal error: %v", err)
			return
		}
		switch p.Type {
		case "text":
			goclipboard.Write(goclipboard.FmtText, []byte(p.Content))
		case "image":
			goclipboard.Write(goclipboard.FmtImage, p.Image)
		case "file_offer":
			log.Printf("[sync] received file offer: %s (%d bytes)", p.FileName, p.FileSize)
			if s.onFileOffer != nil {
				addr := peerAddrFromPeers(s.manager.GetPeers())
				s.onFileOffer(p.OfferID, p.FileName, p.FileSize, addr)
			}
		case "file_accept":
			log.Printf("[sync] received file accept: %s", p.OfferID)
			if s.onFileAccept != nil {
				s.onFileAccept(p.OfferID)
			}
		}
	})

	if err := s.manager.Start(); err != nil {
		return err
	}
	s.startNetworkWatch()
	return nil
}

// Stop shuts down discovery, the network watcher, and listening.
func (s *SyncService) Stop() {
	if s.watchCancel != nil {
		s.watchCancel()
		s.watchCancel = nil
	}
	s.manager.Stop()
}

// BroadcastText sends text content to all connected peers.
func (s *SyncService) BroadcastText(text string) {
	p := wirePayload{Type: "text", Content: text}
	data, err := json.Marshal(p)
	if err != nil {
		log.Printf("[sync] marshal error: %v", err)
		return
	}
	s.manager.Broadcast(data)
}

// BroadcastImage sends image data to all connected peers.
func (s *SyncService) BroadcastImage(img []byte) {
	p := wirePayload{Type: "image", Image: img}
	data, err := json.Marshal(p)
	if err != nil {
		log.Printf("[sync] marshal error: %v", err)
		return
	}
	s.manager.Broadcast(data)
}

// DecryptFailCount returns consecutive decrypt failures from peers.
func (s *SyncService) DecryptFailCount() int {
	if s.manager == nil {
		return 0
	}
	return s.manager.DecryptFailCount()
}

// PeerCount returns the number of connected peers.
func (s *SyncService) PeerCount() int {
	return s.manager.PeerCount()
}

// GetPeers returns a snapshot of connected peers.
func (s *SyncService) GetPeers() []map[string]string {
	peers := s.manager.GetPeers()
	result := make([]map[string]string, len(peers))
	for i, p := range peers {
		result[i] = map[string]string{
			"id":       p.ID,
			"hostname": p.Hostname,
			"addr":     p.Addr,
			"lastSeen": p.LastSeen.Format("15:04:05"),
		}
	}
	return result
}

// BroadcastFileOffer sends a file offer to all connected peers.
func (s *SyncService) BroadcastFileOffer(offerID, fileName string, fileSize int64) {
	p := wirePayload{
		Type:     "file_offer",
		OfferID:  offerID,
		FileName: fileName,
		FileSize: fileSize,
	}
	data, err := json.Marshal(p)
	if err != nil {
		log.Printf("[sync] marshal file offer: %v", err)
		return
	}
	s.manager.Broadcast(data)
}

// peerAddrFromPeers returns the address of the first peer, or "" if empty.
// Used to determine which peer sent a file offer.
func peerAddrFromPeers(peers []sync.Peer) string {
	for _, p := range peers {
		return p.Addr
	}
	return ""
}
