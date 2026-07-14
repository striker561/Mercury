package sync

import (
	"context"
	"log"
	"sync"

	"mercury/app/backend/crypto"
	"mercury/app/backend/transport"
)

// OnReceiveCallback is called when a decrypted clipboard payload arrives.
type OnReceiveCallback func(payload []byte)

// Manager wires mDNS discovery, TCP transport, and crypto for clipboard sync.
// Create with NewManager, then Start/Stop.
type Manager struct {
	key      []byte
	peerMap  *PeerMap
	incoming chan []byte

	mu           sync.Mutex
	running      bool
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	onReceive    OnReceiveCallback
	decryptFails int

	// OnMessage, if set, is called for every message the listener receives
	// that is NOT MsgClipboard.  The app layer uses this to route file
	// chunks to the transfer manager without coupling sync to transfer.
	OnMessage func(msgType byte, payload []byte)
}

// NewManager creates a Manager that derives its encryption key from the
// given passphrase.  Call Start to begin discovery and listening.
func NewManager(passphrase string) *Manager {
	return &Manager{
		key:      crypto.DeriveKey(passphrase),
		peerMap:  NewPeerMap(),
		incoming: make(chan []byte, 10),
	}
}

// SetOnReceive registers a callback for every decrypted clipboard payload.
func (m *Manager) SetOnReceive(cb OnReceiveCallback) {
	m.onReceive = cb
}

// Start begins LAN discovery, TCP listening, and the event loop.
func (m *Manager) Start() error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	added := make(chan Peer, 10)

	if err := Announce(ctx, transport.Port, ""); err != nil {
		cancel()
		return err
	}

	go func() {
		// Sync owns the TCP listener but routes non-clipboard messages
		// (file chunks) to the app layer via OnMessage.
		if err := transport.Listen(ctx, transport.Port, func(msgType byte, payload []byte) {
			if msgType == transport.MsgClipboard {
				select {
				case m.incoming <- payload:
				default:
					log.Printf("[sync] dropped incoming (channel full)")
				}
				return
			}
			if m.OnMessage != nil {
				m.OnMessage(msgType, payload)
			}
		}); err != nil {
			if ctx.Err() == nil {
				log.Printf("[sync] listen error: %v", err)
			}
		}
	}()

	go func() {
		if err := Browse(ctx, added); err != nil {
			if ctx.Err() == nil {
				log.Printf("[sync] browse error: %v", err)
			}
		}
	}()

	m.wg.Add(1)
	go m.eventLoop(ctx, added)

	m.mu.Lock()
	m.running = true
	m.cancel = cancel
	m.mu.Unlock()
	return nil
}

// Stop shuts down the manager, cancelling all goroutines and closing peer
// connections.  Safe to call multiple times or when not running.
func (m *Manager) Stop() {
	m.mu.Lock()
	cancel := m.cancel
	running := m.running
	if running {
		m.running = false
	}
	m.mu.Unlock()

	if !running {
		return
	}

	cancel()
	m.wg.Wait()
	m.peerMap.Stop()
}

// Restart stops the manager with a new passphrase and starts it again.
func (m *Manager) Restart(passphrase string) error {
	m.Stop()

	m.mu.Lock()
	m.key = crypto.DeriveKey(passphrase)
	m.peerMap = NewPeerMap()
	m.incoming = make(chan []byte, 10)
	m.decryptFails = 0
	m.mu.Unlock()

	return m.Start()
}

// Broadcast encrypts the payload and sends it to all currently known peers
// in a background goroutine.  The clipboard handler returns instantly;
// failed sends are logged and silently skipped.
func (m *Manager) Broadcast(payload []byte) {
	m.mu.Lock()
	key := m.key
	running := m.running
	m.mu.Unlock()

	if !running {
		return
	}

	if len(payload) > transport.MaxClipboardSize {
		log.Printf("[sync] skipping oversized payload (%d > %d bytes)", len(payload), transport.MaxClipboardSize)
		return
	}

	ciphertext, err := crypto.Encrypt(payload, key)
	if err != nil {
		log.Printf("[sync] encrypt failed: %v", err)
		return
	}

	peers := m.peerMap.GetPeers()
	if len(peers) == 0 {
		return
	}

	// Send to all peers concurrently — one slow peer should not delay the rest.
	var wg sync.WaitGroup
	for _, p := range peers {
		wg.Add(1)
		p := p // capture loop variable
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), transport.SendTimeout)
			defer cancel()
			if err := transport.Send(ctx, p.Addr, ciphertext); err != nil {
				log.Printf("[sync] send to %s (%s) failed: %v", p.ID, p.Addr, err)
				if m.peerMap.RecordFailure(p.ID) {
					log.Printf("[sync] evicted peer %s (send failures)", p.ID)
				}
			} else {
				m.peerMap.ResetFailures(p.ID)
			}
		}()
	}
	wg.Wait()
}

// PeerCount returns the number of currently known peers.
func (m *Manager) PeerCount() int {
	return m.peerMap.Len()
}

// GetPeers returns a snapshot of all known peers.
func (m *Manager) GetPeers() []Peer {
	return m.peerMap.GetPeers()
}

// Running returns whether the manager is currently started.
func (m *Manager) Running() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

// DecryptFailCount returns consecutive clipboard decrypt failures (wrong passphrase).
func (m *Manager) DecryptFailCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.decryptFails
}

// eventLoop decrypts clipboard payloads and dispatches to onReceive.
// File chunks are handled entirely by the app layer via OnMessage —
// this event loop knows nothing about them.
func (m *Manager) eventLoop(ctx context.Context, added <-chan Peer) {
	defer m.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case ciphertext := <-m.incoming:
			decrypted, err := crypto.Decrypt(ciphertext, m.key)
			if err != nil {
				m.mu.Lock()
				m.decryptFails++
				m.mu.Unlock()
				log.Printf("[sync] decrypt failed (wrong key?): %v", err)
			} else {
				m.mu.Lock()
				m.decryptFails = 0
				m.mu.Unlock()
				if m.onReceive != nil {
					m.onReceive(decrypted)
				}
			}
		case peer := <-added:
			m.peerMap.AddOrUpdate(peer.ID, peer.Addr)
		}
	}
}
