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
	"log"
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
	ID        string `json:"id"`
	FileName  string `json:"file_name"`
	FileSize  int64  `json:"file_size"`
	Received  int64  `json:"received"`
	Speed     int64  `json:"speed"`   // bytes/sec, 0 when idle
	Status    Status `json:"status"`
}

// ─── Manager ─────────────────────────────────────────────────────────

// Manager handles concurrent file offers and transfers.
type Manager struct {
	key       []byte
	mu        sync.Mutex
	offers    map[string]*Offer // incoming offers (from network)
	outgoing  map[string]string // our offers: offerID → filePath
	transfers map[string]*Progress
	cancel    map[string]chan struct{} // tid → close to cancel
	nextID    atomic.Int64

	chunkBuf chan []byte // decrypted chunks from sync listener

	// OnOffer is called when a file offer arrives from the network.
	OnOffer func(o Offer)

	// OnAccept is called when a remote peer accepts one of our offers.
	OnAccept func(offerID string)
}

// NewManager creates a transfer manager that uses the given encryption key.
func NewManager(key []byte) *Manager {
	return &Manager{
		key:       key,
		offers:    make(map[string]*Offer),
		outgoing:  make(map[string]string),
		transfers: make(map[string]*Progress),
		cancel:    make(map[string]chan struct{}),
		chunkBuf:  make(chan []byte, 20),
	}
}

// ChunkChan returns the channel where file chunks should be sent.
// The app wires this by feeding decrypted chunks from the sync manager.
func (m *Manager) ChunkChan() chan<- []byte {
	return m.chunkBuf
}

// NewOfferID returns a random hex identifier for offers and transfers.
func (m *Manager) NewOfferID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// IncomingOffer stores an offer from the network and fires OnOffer.
func (m *Manager) IncomingOffer(fileName string, fileSize int64, peerAddr string) Offer {
	id := m.NewOfferID()
	o := Offer{ID: id, FileName: fileName, FileSize: fileSize, PeerAddr: peerAddr}
	m.mu.Lock()
	m.offers[id] = &o
	m.mu.Unlock()
	log.Printf("[transfer] incoming offer: %s (%d bytes)", fileName, fileSize)
	if m.OnOffer != nil {
		m.OnOffer(o)
	}
	return o
}

// IncomingOfferWithID stores an offer using the sender's offer ID (no new ID generated).
func (m *Manager) IncomingOfferWithID(offerID, fileName string, fileSize int64, peerAddr string) Offer {
	o := Offer{ID: offerID, FileName: fileName, FileSize: fileSize, PeerAddr: peerAddr}
	m.mu.Lock()
	m.offers[offerID] = &o
	m.mu.Unlock()
	log.Printf("[transfer] incoming offer (sender's ID): %s (%d bytes)", fileName, fileSize)
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

	tid := m.NewOfferID()
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
	log.Printf("[transfer] rejected offer %s", offerID)
}

// StoreOutgoing remembers an offer we broadcast so we can send the file when accepted.
func (m *Manager) StoreOutgoing(offerID, filePath string) {
	m.mu.Lock()
	m.outgoing[offerID] = filePath
	m.mu.Unlock()
}

// AcceptNotification is called when a remote peer accepts one of our offers.
func (m *Manager) AcceptNotification(offerID string) string {
	m.mu.Lock()
	fp := m.outgoing[offerID]
	delete(m.outgoing, offerID)
	m.mu.Unlock()
	return fp
}

// SendFile sends a file to a peer. Returns a transfer ID.
func (m *Manager) SendFile(peerAddr, filePath string) (string, error) {
	fi, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("transfer: %w", err)
	}
	tid := m.NewOfferID()
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

// CancelTransfer signals a running transfer (send or receive) to stop.
// The partial file is cleaned up on the receiving side.
func (m *Manager) CancelTransfer(tid string) {
	m.mu.Lock()
	c, ok := m.cancel[tid]
	m.mu.Unlock()
	if ok {
		close(c)
	}
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

// updateSpeed stores the current transfer speed in bytes/sec.
func (m *Manager) updateSpeed(tid string, speed int64) {
	m.mu.Lock()
	if p, ok := m.transfers[tid]; ok {
		p.Speed = speed
	}
	m.mu.Unlock()
}
