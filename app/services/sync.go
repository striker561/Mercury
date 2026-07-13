package services

import (
	"encoding/json"
	"log"

	"mercury/app/backend/sync"

	goclipboard "golang.design/x/clipboard"
)

// wirePayload is the JSON structure sent over the network for clipboard sync.
type wirePayload struct {
	Type    string `json:"type"`    // "text" or "image"
	Content string `json:"content"` // text content (for text type)
	Image   []byte `json:"image"`   // PNG bytes (for image type)
}

// SyncService manages LAN peer discovery, clipboard broadcast, and receive.
type SyncService struct {
	manager *sync.Manager
}

// NewSyncService creates a sync service with the given passphrase.
func NewSyncService(passphrase string) *SyncService {
	return &SyncService{
		manager: sync.NewManager(passphrase),
	}
}

// Start begins mDNS discovery and TCP listening.
// Received clipboard data is automatically written to the OS clipboard.
func (s *SyncService) Start() error {
	// Wire receive path: network -> OS clipboard
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
