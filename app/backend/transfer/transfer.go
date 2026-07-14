// Package transfer handles peer-to-peer file transfer.
//
// Flow:
//  1. Offer (name+size+id) broadcast via sync channel
//  2. User accepts → receiver goroutine starts, reads from chunkBuf
//  3. Sender encrypts 256 KiB chunks, sends each via sync.SendMsg(MsgFileChunk)
//  4. Sync listener demuxes by type byte, decrypts, pushes to chunkBuf
//  5. Receiver writes chunks to disk
//
// Wire: same port (47821), same AES-256-GCM key as clipboard sync,
// but with MsgFileChunk type byte so the listener routes to us.
package transfer

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

const (
	chunkSize   = 256 * 1024 // 256 KiB per chunk
	dialTimeout = 5 * time.Second
)

// ─── Types ───────────────────────────────────────────────────────────

// Offer represents a pending incoming file offer.
type Offer struct {
	ID       string `json:"id"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	PeerAddr string `json:"peer_addr"`
}

// Status is the current state of a file transfer.
type Status string

const (
	StatusOffered   Status = "offered"
	StatusAccepting Status = "accepting"
	StatusSending   Status = "sending"
	StatusReceiving Status = "receiving"
	StatusDone      Status = "done"
	StatusFailed    Status = "failed"
)

// Progress contains live transfer state for the frontend.
type Progress struct {
	ID       string `json:"id"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	Received int64  `json:"received"`
	Status   Status `json:"status"`
}

// ─── Manager ─────────────────────────────────────────────────────────

// Manager handles concurrent file offers and transfers.
type Manager struct {
	key       []byte
	mu        sync.Mutex
	offers    map[string]*Offer
	transfers map[string]*Progress
	nextID    atomic.Int64

	// chunkBuf receives encrypted file chunks from the sync manager.
	// receiveFile reads from it and writes to disk.
	chunkBuf chan []byte

	// OnOffer is called when a file offer arrives from the network.
	OnOffer func(o Offer)
}

// NewManager creates a transfer manager that uses the given encryption key.
func NewManager(key []byte) *Manager {
	return &Manager{
		key:       key,
		offers:    make(map[string]*Offer),
		transfers: make(map[string]*Progress),
		chunkBuf:  make(chan []byte, 20),
	}
}

// ChunkChan returns the channel where file chunks should be sent.
// The app wires this by feeding decrypted chunks from the sync manager.
func (m *Manager) ChunkChan() chan<- []byte {
	return m.chunkBuf
}

// newID returns a random hex identifier for offers and transfers.
func (m *Manager) newID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// IncomingOffer stores an offer from the network and fires OnOffer.
func (m *Manager) IncomingOffer(fileName string, fileSize int64, peerAddr string) Offer {
	id := m.newID()
	o := Offer{ID: id, FileName: fileName, FileSize: fileSize, PeerAddr: peerAddr}
	m.mu.Lock()
	m.offers[id] = &o
	m.mu.Unlock()
	if m.OnOffer != nil {
		m.OnOffer(o)
	}
	return o
}

// PendingOffers returns all offers that haven't been acted on.
func (m *Manager) PendingOffers() []Offer {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Offer, 0, len(m.offers))
	for _, o := range m.offers {
		out = append(out, *o)
	}
	return out
}

// AcceptOffer starts receiving the accepted file. Returns a transfer ID.
func (m *Manager) AcceptOffer(offerID, saveDir string) (string, error) {
	m.mu.Lock()
	o, ok := m.offers[offerID]
	if !ok {
		m.mu.Unlock()
		return "", fmt.Errorf("offer %s not found", offerID)
	}
	delete(m.offers, offerID)
	m.mu.Unlock()

	tid := m.newID()
	p := &Progress{ID: tid, FileName: o.FileName, FileSize: o.FileSize, Status: StatusAccepting}
	m.mu.Lock()
	m.transfers[tid] = p
	m.mu.Unlock()

	go m.receiveFile(tid, o, saveDir)
	return tid, nil
}

// RejectOffer removes a pending offer.
func (m *Manager) RejectOffer(offerID string) {
	m.mu.Lock()
	delete(m.offers, offerID)
	m.mu.Unlock()
}

// SendFile sends a file to a peer. Returns a transfer ID.
func (m *Manager) SendFile(peerAddr, filePath string) (string, error) {
	fi, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("transfer: %w", err)
	}
	tid := m.newID()
	p := &Progress{ID: tid, FileName: filepath.Base(filePath), FileSize: fi.Size(), Status: StatusSending}
	m.mu.Lock()
	m.transfers[tid] = p
	m.mu.Unlock()

	go m.sendFile(tid, peerAddr, filePath, fi.Size())
	return tid, nil
}

// Progress returns a copy of a transfer's progress, or nil.
func (m *Manager) Progress(tid string) *Progress {
	m.mu.Lock()
	defer m.mu.Unlock()
	p := m.transfers[tid]
	if p == nil {
		return nil
	}
	cp := *p
	return &cp
}

// AllProgress returns progress for active transfers (omits completed).
func (m *Manager) AllProgress() []Progress {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Progress, 0, len(m.transfers))
	for _, p := range m.transfers {
		if p.Status == StatusDone {
			continue
		}
		out = append(out, *p)
	}
	return out
}

// updateStatus is a helper used by sendFile / receiveFile.
func (m *Manager) updateStatus(tid string, s Status, received int64) {
	m.mu.Lock()
	if p, ok := m.transfers[tid]; ok {
		p.Status = s
		p.Received = received
	}
	m.mu.Unlock()
}
