package sync

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
)

func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port
}

// A restart must not race the old listener's socket close.
func TestManagerStartStopStart(t *testing.T) {
	m := NewManager("test-passphrase", "test-device")
	m.port = freePort(t)

	for i := 0; i < 3; i++ {
		if err := m.Start(); err != nil {
			t.Fatalf("start %d: %v", i, err)
		}
		if !m.Running() {
			t.Fatalf("not running after start %d", i)
		}
		m.Stop()
		if m.Running() {
			t.Fatalf("still running after stop %d", i)
		}
	}
}

func TestManagerRestartDiscoveryKeepsListener(t *testing.T) {
	m := NewManager("test-passphrase", "test-device")
	m.port = freePort(t)

	if err := m.Start(); err != nil {
		t.Fatal(err)
	}
	defer m.Stop()

	if err := m.RestartDiscovery(); err != nil {
		t.Fatalf("restart discovery: %v", err)
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", m.boundPort), time.Second)
	if err != nil {
		t.Fatalf("listener dropped after discovery restart: %v", err)
	}
	conn.Close()
}

func TestManagerConcurrentStart(t *testing.T) {
	m := NewManager("test-passphrase", "test-device")
	m.port = freePort(t)

	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- m.Start()
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent start: %v", err)
		}
	}
	if !m.Running() {
		t.Fatal("not running after concurrent start")
	}
	m.Stop()
}

func TestProbePeerMarksReachable(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	m := NewManager("test-passphrase", "test-device")
	m.peerMap.AddOrUpdatePeer(Peer{ID: "peer1", Hostname: "peer1", Addr: ln.Addr().String()})
	m.probePeer(context.Background(), m.peerMap.GetPeers()[0])

	peers := m.peerMap.GetPeers()
	if len(peers) != 1 || !peers[0].Reachable {
		t.Fatalf("peer not marked reachable: %+v", peers)
	}
}

func TestProbePeerEvictsUnreachable(t *testing.T) {
	m := NewManager("test-passphrase", "test-device")
	p := Peer{ID: "peer1", Hostname: "peer1", Addr: "127.0.0.1:1"}
	m.peerMap.AddOrUpdatePeer(p)

	for i := 0; i < maxFailCount; i++ {
		m.probePeer(context.Background(), p)
	}

	if m.peerMap.Len() != 0 {
		t.Fatalf("peer should be evicted, still have %d", m.peerMap.Len())
	}
}
