package sync

import (
	"context"
	"fmt"
	"log"
	"net"
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
	deviceID string
	peerMap  *PeerMap
	incoming chan []byte
	port     int

	mu           sync.Mutex
	running      bool
	lastErr      string
	decryptFails int

	// lifecycleMu serialises Start, Stop, Restart and RestartDiscovery.
	lifecycleMu  sync.Mutex
	tcpCancel    context.CancelFunc
	tcpWG        sync.WaitGroup
	discCancel   context.CancelFunc
	discWG       sync.WaitGroup
	stopAnnounce func()
	added        chan Peer
	boundPort    int

	onReceive OnReceiveCallback

	// OnMessage, if set, is called for every message the listener receives
	// that is NOT MsgClipboard.  The app layer uses this to route file
	// chunks to the transfer manager without coupling sync to transfer.
	OnMessage func(msgType byte, payload []byte)
}

// NewManager creates a Manager that derives its encryption key from the
// given passphrase.  deviceID is this machine's stable identity, advertised
// in mDNS TXT records so peers can dedup us across hostname changes.  Call
// Start to begin discovery and listening.
func NewManager(passphrase, deviceID string) *Manager {
	return &Manager{
		key:      crypto.DeriveKey(passphrase),
		deviceID: deviceID,
		peerMap:  NewPeerMap(),
		incoming: make(chan []byte, 10),
		port:     transport.Port,
	}
}

// SetOnReceive registers a callback for every decrypted clipboard payload.
func (m *Manager) SetOnReceive(cb OnReceiveCallback) {
	m.onReceive = cb
}

// Start begins LAN discovery, TCP listening, and the event loop.
func (m *Manager) Start() error {
	m.lifecycleMu.Lock()
	defer m.lifecycleMu.Unlock()
	return m.startLocked()
}

func (m *Manager) startLocked() error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()

	listener, err := transport.Bind(m.port)
	if err != nil {
		m.setError(err)
		return err
	}

	port := m.port
	if tcp, ok := listener.Addr().(*net.TCPAddr); ok {
		port = tcp.Port
	}

	ctx, cancel := context.WithCancel(context.Background())
	added := make(chan Peer, 10)

	m.mu.Lock()
	m.tcpCancel = cancel
	m.mu.Unlock()
	m.boundPort = port
	m.tcpWG.Add(2)
	go func() {
		defer m.tcpWG.Done()
		if err := transport.Serve(ctx, listener, m.handleMessage); err != nil && ctx.Err() == nil {
			log.Printf("[sync] listen error: %v", err)
		}
	}()
	go func() {
		defer m.tcpWG.Done()
		m.eventLoop(ctx, added)
	}()

	if err := m.startDiscoveryLocked(added, port); err != nil {
		cancel()
		m.tcpWG.Wait()
		m.mu.Lock()
		m.tcpCancel = nil
		m.mu.Unlock()
		m.setError(err)
		return err
	}

	m.mu.Lock()
	m.running = true
	m.lastErr = ""
	m.mu.Unlock()
	return nil
}

func (m *Manager) handleMessage(msgType byte, payload []byte) {
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
}

func (m *Manager) startDiscoveryLocked(added chan Peer, port int) error {
	ctx, cancel := context.WithCancel(context.Background())
	stop, err := Announce(ctx, port, "", m.deviceID)
	if err != nil {
		cancel()
		return err
	}

	m.discCancel = cancel
	m.stopAnnounce = stop
	m.added = added
	m.discWG.Add(1)
	go func() {
		defer m.discWG.Done()
		if err := Browse(ctx, added, m.deviceID); err != nil && ctx.Err() == nil {
			log.Printf("[sync] browse error: %v", err)
		}
	}()
	return nil
}

func (m *Manager) stopDiscoveryLocked() {
	if m.discCancel != nil {
		m.discCancel()
		m.discCancel = nil
	}
	m.discWG.Wait()
	if m.stopAnnounce != nil {
		m.stopAnnounce()
		m.stopAnnounce = nil
	}
}

// Stop shuts down the manager, cancelling all goroutines and closing peer
// connections.  Safe to call multiple times or when not running.
func (m *Manager) Stop() {
	m.lifecycleMu.Lock()
	defer m.lifecycleMu.Unlock()
	m.stopLocked()
}

func (m *Manager) stopLocked() {
	m.mu.Lock()
	m.running = false
	cancel := m.tcpCancel
	m.tcpCancel = nil
	m.mu.Unlock()

	m.stopDiscoveryLocked()
	if cancel != nil {
		cancel()
	}
	m.tcpWG.Wait()
	m.peerMap.Stop()
}

// Restart stops the manager with a new passphrase and starts it again.
func (m *Manager) Restart(passphrase string) error {
	m.lifecycleMu.Lock()
	defer m.lifecycleMu.Unlock()

	m.stopLocked()
	for len(m.incoming) > 0 {
		<-m.incoming
	}

	m.mu.Lock()
	m.key = crypto.DeriveKey(passphrase)
	m.decryptFails = 0
	m.mu.Unlock()
	m.peerMap.Clear()

	return m.startLocked()
}

// RestartDiscovery re-creates the mDNS resolver and registration.  The TCP
// listener and the known peers stay up, so a network change cannot drop an
// in-flight message or force a port rebind.
func (m *Manager) RestartDiscovery() error {
	m.lifecycleMu.Lock()
	defer m.lifecycleMu.Unlock()

	m.mu.Lock()
	running := m.running
	m.mu.Unlock()
	if !running {
		return fmt.Errorf("sync not running")
	}

	m.stopDiscoveryLocked()
	if err := m.startDiscoveryLocked(m.added, m.boundPort); err != nil {
		m.setError(err)
		return err
	}
	m.setError(nil)
	return nil
}

// Resync restarts discovery without changing the encryption key or dropping
// known peers.
func (m *Manager) Resync() error {
	return m.RestartDiscovery()
}

func (m *Manager) setError(err error) {
	m.mu.Lock()
	if err == nil {
		m.lastErr = ""
	} else {
		m.lastErr = err.Error()
	}
	m.mu.Unlock()
}

// LastError returns the last lifecycle error, or "" when healthy.
func (m *Manager) LastError() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastErr
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
	for {
		select {
		case <-ctx.Done():
			return
		case ciphertext := <-m.incoming:
			m.mu.Lock()
			key := m.key
			m.mu.Unlock()
			decrypted, err := crypto.Decrypt(ciphertext, key)
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
			m.peerMap.AddOrUpdatePeer(peer)
		}
	}
}
