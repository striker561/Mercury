package sync

import (
	"sort"
	"sync"
	"time"
)

const (
	// maxFailCount is the number of consecutive TCP failures before a peer
	// is evicted.  Reset to 0 on any successful send or heartbeat.
	maxFailCount = 3
)

// Peer represents a Mercury instance discovered on the LAN.
//
// ID is the stable machine ID (announced via mDNS TXT) — this is what the
// peer map is keyed by, so a dual-boot machine (same IP, different hostname
// per OS) stays a single entry.  Hostname is the announced display name and
// may change; it's for the UI only.  Legacy peers that don't announce a
// machine ID use their hostname as ID.
type Peer struct {
	ID        string    `json:"id"`
	Hostname  string    `json:"hostname"`
	Addr      string    `json:"addr"`
	LastSeen  time.Time `json:"lastSeen"`
	failCount int       // consecutive TCP failures (internal, not exported)
}

// PeerMap is a concurrency-safe map of peers.  Peers are only removed when
// RecordFailure reaches maxFailCount — there is no TTL-based eviction.
// Re-discovery via mDNS will re-add evicted peers when they come back.
type PeerMap struct {
	mu    sync.RWMutex
	peers map[string]*Peer
}

// NewPeerMap creates an empty PeerMap.  No background goroutines.
func NewPeerMap() *PeerMap {
	return &PeerMap{
		peers: make(map[string]*Peer),
	}
}

// AddOrUpdate records or refreshes a peer keyed by id.  If a peer with the
// same ID already exists its address and LastSeen are updated and its failure
// count is reset (a new mDNS announcement means the peer is alive).  Hostname
// defaults to id for legacy callers.
func (pm *PeerMap) AddOrUpdate(id, addr string) {
	pm.addOrUpdate(&Peer{ID: id, Addr: addr, Hostname: id})
}

// AddOrUpdatePeer records or refreshes a peer with full metadata, used by the
// mDNS discovery path where the stable machine ID and the display hostname
// can differ (dual-boot).  Updating an existing peer refreshes its hostname,
// so the entry shows whichever OS last announced it.
func (pm *PeerMap) AddOrUpdatePeer(p Peer) {
	pm.addOrUpdate(&p)
}

func (pm *PeerMap) addOrUpdate(p *Peer) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if existing, ok := pm.peers[p.ID]; ok {
		existing.Addr = p.Addr
		existing.Hostname = p.Hostname
		existing.LastSeen = time.Now()
		existing.failCount = 0 // fresh announcement = alive
		return
	}

	p.LastSeen = time.Now()
	pm.peers[p.ID] = p
}

// RecordFailure increments a peer's failure counter.  Returns true when
// the counter reaches maxFailCount, signalling the caller to evict.
func (pm *PeerMap) RecordFailure(id string) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p, ok := pm.peers[id]
	if !ok {
		return false
	}
	p.failCount++
	if p.failCount >= maxFailCount {
		delete(pm.peers, id)
		return true
	}
	return false
}

// ResetFailures resets a peer's failure counter to zero.
func (pm *PeerMap) ResetFailures(id string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if p, ok := pm.peers[id]; ok {
		p.failCount = 0
	}
}

// GetPeers returns a snapshot of all currently known peers, sorted
// alphabetically by ID (which is the announced hostname).  Sorting here keeps
// the peer list stable across the UI's frequent re-fetches — Go map iteration
// order is randomised, and an unstable order made the list jump around.
func (pm *PeerMap) GetPeers() []Peer {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make([]Peer, 0, len(pm.peers))
	for _, p := range pm.peers {
		result = append(result, *p) // copy
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Hostname < result[j].Hostname
	})
	return result
}

// Remove deletes a peer by ID.
func (pm *PeerMap) Remove(id string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.peers, id)
}

// Len returns the number of currently known peers.
func (pm *PeerMap) Len() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.peers)
}

// Stop is a no-op kept for interface compatibility.  No background
// goroutines to shut down.
func (pm *PeerMap) Stop() {}
