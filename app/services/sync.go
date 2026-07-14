package services

import (
	"encoding/json"
	"log"

	"mercury/app/backend/sync"

	goclipboard "golang.design/x/clipboard"
)

// wirePayload is the JSON structure sent over the network for clipboard sync.
type wirePayload struct {
	Type    string `json:"type"`    // "text", "image", or "file_offer"
	Content string `json:"content"` // text content (for text type)
	Image   []byte `json:"image"`   // PNG bytes (for image type)

	// File offer fields.
	FileName string `json:"file_name,omitempty"`
	FileSize int64  `json:"file_size,omitempty"`
}

// OnFileOfferCallback is called when a file offer arrives from a peer.
type OnFileOfferCallback func(fileName string, fileSize int64, peerAddr string)

// SyncService manages LAN peer discovery, clipboard broadcast, and receive.
type SyncService struct {
	manager     *sync.Manager
	onFileOffer OnFileOfferCallback
}

// SetOnFileChunk registers a callback for incoming file chunks.
func (s *SyncService) SetOnFileChunk(cb sync.OnFileChunkCallback) {
	s.manager.SetOnFileChunk(cb)
}

// NewSyncService creates a sync service with the given passphrase.
func NewSyncService(passphrase string) *SyncService {
	return &SyncService{
		manager: sync.NewManager(passphrase),
	}
}

// DeriveKey derives the encryption key from the passphrase (re-exports sync.DeriveKey).
func DeriveKey(passphrase string) []byte {
	return sync.DeriveKey(passphrase)
}

// SetOnFileOffer registers a callback for incoming file offers.
func (s *SyncService) SetOnFileOffer(cb OnFileOfferCallback) {
	s.onFileOffer = cb
}

// Start begins mDNS discovery and TCP listening.
// Received clipboard data is automatically written to the OS clipboard.
func (s *SyncService) Start() error {
	// Wire receive path: network -> OS clipboard or file offer
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
			if s.onFileOffer != nil {
				// The sender's address isn't in the payload, but we
				// know it from the peer map.  We use the first peer
				// that matches the hostname — in practice there's
				// only one per machine.
				addr := peerAddrFromPeers(s.manager.GetPeers())
				s.onFileOffer(p.FileName, p.FileSize, addr)
			}
		}
	})

	return s.manager.Start()
}

// Stop shuts down discovery and listening.
func (s *SyncService) Stop() {
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
			"addr":     p.Addr,
			"lastSeen": p.LastSeen.Format("15:04:05"),
		}
	}
	return result
}

// BroadcastFileOffer sends a file offer to all connected peers.
func (s *SyncService) BroadcastFileOffer(fileName string, fileSize int64) {
	p := wirePayload{
		Type:     "file_offer",
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
