// Package transfer handles peer-to-peer file transfer.
//
// Flow:
//  1. Offer (name+size+id) broadcast via sync channel
//  2. User accepts → receiver goroutine starts on a per-offer chunk channel
//  3. Sender dials one TCP connection, encrypts 256 KiB chunks (prefixed with
//     the offer ID), and streams them via transport.WriteFrame(MsgFileChunk)
//  4. Sync listener demuxes by type byte, decrypts, routes chunks by offer ID
//  5. Sender sends MsgFileEnd on the same connection; receiver writes to disk
//
// Wire: same port (47821), same AES-256-GCM key as clipboard sync.
// File payloads are tagged with the 16-byte offer ID so concurrent transfers
// to the same machine do not share a global chunk buffer.
package transfer

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"mercury/app/backend/transport"
)

const (
	chunkSize    = 256 * 1024 // 256 KiB per chunk
	dialTimeout  = 5 * time.Second
	writeTimeout = 30 * time.Second

	maxTerminalTransfers = 50
	maxOutgoingOffers    = 50
	outgoingOfferTTL     = 10 * time.Minute
)

// ─── Types ───────────────────────────────────────────────────────────

// Offer represents a pending incoming file offer.
type Offer struct {
	ID        string    `json:"id"`
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
	PeerAddr  string    `json:"peer_addr"`
	CreatedAt time.Time `json:"created_at"` // arrival order — used to stabilise the UI list
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

type outgoingOffer struct {
	path string
	at   time.Time
}

type receiveStream struct {
	chunks chan []byte
	end    chan struct{}
	done   chan struct{}
}

// Progress contains live transfer state for the frontend.
type Progress struct {
	ID        string    `json:"id"`
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
	Received  int64     `json:"received"`
	Speed     int64     `json:"speed"` // bytes/sec, 0 when idle
	Status    Status    `json:"status"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"` // arrival order — used to stabilise the UI list
}

// ─── Manager ─────────────────────────────────────────────────────────

// Manager handles concurrent file offers and transfers.
type Manager struct {
	key       []byte
	mu        sync.Mutex
	offers    map[string]*Offer        // incoming offers (from network)
	outgoing  map[string]outgoingOffer // our offers: offerID → file path + time
	transfers map[string]*Progress
	cancel    map[string]chan struct{} // tid → close to cancel
	nextID    atomic.Int64

	receiveCh map[string]*receiveStream // offer ID → active receive

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
		outgoing:  make(map[string]outgoingOffer),
		transfers: make(map[string]*Progress),
		cancel:    make(map[string]chan struct{}),
		receiveCh: make(map[string]*receiveStream),
	}
}

// OnWireMessage routes a decrypted file-transfer frame to the active receive
// for its offer ID. Called from the app layer after sync demuxes MsgFileChunk
// or MsgFileEnd off the shared TCP listener.
func (m *Manager) OnWireMessage(msgType byte, payload []byte) {
	switch msgType {
	case transport.MsgFileChunk:
		offerID, err := offerIDFromPayload(payload)
		if err != nil {
			log.Printf("[transfer] chunk: %v", err)
			return
		}
		chunk := payload[offerIDLen:]

		m.mu.Lock()
		stream := m.receiveCh[offerID]
		m.mu.Unlock()
		if stream == nil {
			return
		}

		// Blocking send is backpressure: the TCP reader waits for the disk.
		select {
		case stream.chunks <- chunk:
		case <-stream.done:
		}

	case transport.MsgFileEnd:
		if len(payload) != offerIDLen {
			log.Printf("[transfer] end marker: bad payload length %d", len(payload))
			return
		}
		offerID := string(payload)

		m.mu.Lock()
		stream := m.receiveCh[offerID]
		delete(m.receiveCh, offerID)
		m.mu.Unlock()
		if stream == nil {
			return
		}
		close(stream.end)
	}
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
	o := Offer{ID: id, FileName: fileName, FileSize: fileSize, PeerAddr: peerAddr, CreatedAt: time.Now()}
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
	o := Offer{ID: offerID, FileName: fileName, FileSize: fileSize, PeerAddr: peerAddr, CreatedAt: time.Now()}
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
// PendingOffers returns all offers that haven't been acted on, sorted by
// arrival time (oldest first) with ID as a deterministic tie-break.  Sorting
// here is what keeps the frontend list stable — Go map iteration is
// randomised, and the UI re-fetches on every progress event.
func (m *Manager) PendingOffers() []Offer {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Offer, 0, len(m.offers))
	for _, o := range m.offers {
		out = append(out, *o)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

// OfferPeer returns the address that sent a pending offer.
func (m *Manager) OfferPeer(offerID string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if o, ok := m.offers[offerID]; ok {
		return o.PeerAddr
	}
	return ""
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
	p := &Progress{ID: tid, FileName: o.FileName, FileSize: o.FileSize, Status: StatusAccepting, CreatedAt: time.Now()}
	m.mu.Lock()
	m.transfers[tid] = p
	m.pruneTerminalLocked()
	m.mu.Unlock()

	stream := &receiveStream{
		chunks: make(chan []byte, 20),
		end:    make(chan struct{}),
		done:   make(chan struct{}),
	}
	m.mu.Lock()
	m.receiveCh[offerID] = stream
	m.mu.Unlock()

	go m.receiveFile(tid, offerID, o, saveDir, stream)
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
	m.outgoing[offerID] = outgoingOffer{path: filePath, at: time.Now()}
	m.pruneOutgoingLocked()
	m.mu.Unlock()
}

// AcceptNotification is called when a remote peer accepts one of our offers.
func (m *Manager) AcceptNotification(offerID string) string {
	m.mu.Lock()
	o, ok := m.outgoing[offerID]
	delete(m.outgoing, offerID)
	m.mu.Unlock()
	if !ok {
		return ""
	}
	return o.path
}

func (m *Manager) pruneTerminalLocked() {
	var terminal []*Progress
	for _, p := range m.transfers {
		if p.Status == StatusDone || p.Status == StatusFailed {
			terminal = append(terminal, p)
		}
	}
	if len(terminal) <= maxTerminalTransfers {
		return
	}
	sort.Slice(terminal, func(i, j int) bool {
		return terminal[i].CreatedAt.Before(terminal[j].CreatedAt)
	})
	for _, p := range terminal[:len(terminal)-maxTerminalTransfers] {
		delete(m.transfers, p.ID)
	}
}

func (m *Manager) pruneOutgoingLocked() {
	cutoff := time.Now().Add(-outgoingOfferTTL)
	for id, o := range m.outgoing {
		if o.at.Before(cutoff) {
			delete(m.outgoing, id)
		}
	}
	if len(m.outgoing) <= maxOutgoingOffers {
		return
	}

	type entry struct {
		id string
		at time.Time
	}
	entries := make([]entry, 0, len(m.outgoing))
	for id, o := range m.outgoing {
		entries = append(entries, entry{id: id, at: o.at})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].at.Before(entries[j].at)
	})
	for _, e := range entries[:len(entries)-maxOutgoingOffers] {
		delete(m.outgoing, e.id)
	}
}

// SendFile sends a file to a peer. Returns a transfer ID.
func (m *Manager) SendFile(peerAddr, filePath string) (string, error) {
	fi, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("transfer: %w", err)
	}
	tid := m.NewOfferID()
	p := &Progress{ID: tid, FileName: filepath.Base(filePath), FileSize: fi.Size(), Status: StatusSending, CreatedAt: time.Now()}
	m.mu.Lock()
	m.transfers[tid] = p
	m.pruneTerminalLocked()
	m.mu.Unlock()

	offerID := m.NewOfferID()
	go m.sendFile(tid, offerID, peerAddr, filePath, fi.Size())
	return tid, nil
}

// SendFileForOffer sends a file for a specific offer ID (after the peer accepted).
func (m *Manager) SendFileForOffer(offerID, peerAddr, filePath string) (string, error) {
	fi, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("transfer: %w", err)
	}
	tid := m.NewOfferID()
	p := &Progress{ID: tid, FileName: filepath.Base(filePath), FileSize: fi.Size(), Status: StatusSending, CreatedAt: time.Now()}
	m.mu.Lock()
	m.transfers[tid] = p
	m.pruneTerminalLocked()
	m.mu.Unlock()

	go m.sendFile(tid, offerID, peerAddr, filePath, fi.Size())
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

// HasActiveTransfers reports whether a send, receive or accept is in flight.
func (m *Manager) HasActiveTransfers() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range m.transfers {
		switch p.Status {
		case StatusAccepting, StatusSending, StatusReceiving:
			return true
		}
	}
	return false
}

// AllProgress returns progress for active transfers (omits completed), sorted
// by arrival time (oldest first) with ID as a deterministic tie-break.  Same
// reasoning as PendingOffers: stable order across UI re-fetches.
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
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

// CancelTransfer signals a running transfer (send or receive) to stop.
// The partial file is cleaned up on the receiving side.
func (m *Manager) CancelTransfer(tid string) {
	m.mu.Lock()
	c, ok := m.cancel[tid]
	if ok {
		delete(m.cancel, tid)
	}
	m.mu.Unlock()
	if ok {
		close(c)
	}
}

// DismissTransfer removes a finished transfer from the dashboard. Active
// transfers are left alone; cancel them first.
func (m *Manager) DismissTransfer(tid string) {
	m.mu.Lock()
	if p, ok := m.transfers[tid]; ok && (p.Status == StatusDone || p.Status == StatusFailed) {
		delete(m.transfers, tid)
	}
	m.mu.Unlock()
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

func (m *Manager) fail(tid, reason string) {
	m.mu.Lock()
	if p, ok := m.transfers[tid]; ok {
		p.Status = StatusFailed
		p.Speed = 0
		p.Error = reason
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
