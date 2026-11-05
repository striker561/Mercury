package sync

import (
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
