package sync

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

// synthetic ifaces with an impossible index so net.Interface.Addrs() returns
// no addresses — the fingerprint then depends only on Name/MTU, which is all
// the tests vary.  Deterministic and OS-independent.
func iface(name string, mtu int) net.Interface {
	return net.Interface{
		Name:  name,
		MTU:   mtu,
		Flags: net.FlagUp | net.FlagMulticast,
	}
}

func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condition not met within timeout")
}

func TestNetworkWatcherResyncsOnInterfaceChange(t *testing.T) {
	net1 := []net.Interface{iface("eth0", 1500)}
	net2 := []net.Interface{iface("wlan0", 1400)}

	var current atomic.Value
	current.Store(net1)
	list := func() ([]net.Interface, error) {
		return current.Load().([]net.Interface), nil
	}

	w := NewNetworkWatcher()
	w.pollInterval = 10 * time.Millisecond
	w.noPeerTimeout = time.Hour // not testing the timeout here
	w.listIfaces = list

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var resyncs atomic.Int32
	done := make(chan struct{})
	go func() {
		defer close(done)
		w.Start(ctx, func() int { return 1 }, func() { resyncs.Add(1) })
	}()

	// Let the watcher take its first reading, then switch networks.
	time.Sleep(30 * time.Millisecond)
	current.Store(net2)

	waitFor(t, 2*time.Second, func() bool { return resyncs.Load() > 0 })
	cancel()
	<-done
}

func TestNetworkWatcherRetriesWhenNoPeers(t *testing.T) {
	w := NewNetworkWatcher()
	w.pollInterval = 10 * time.Millisecond
	w.noPeerTimeout = 25 * time.Millisecond
	w.listIfaces = func() ([]net.Interface, error) { return nil, nil }

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var resyncs atomic.Int32
	done := make(chan struct{})
	go func() {
		defer close(done)
		w.Start(ctx, func() int { return 0 }, func() { resyncs.Add(1) })
	}()

	// Zero peers for longer than noPeerTimeout must trigger a retry.
	waitFor(t, 2*time.Second, func() bool { return resyncs.Load() > 0 })
	cancel()
	<-done
}

func TestNetworkWatcherNoResyncWithPeers(t *testing.T) {
	w := NewNetworkWatcher()
	w.pollInterval = 10 * time.Millisecond
	w.noPeerTimeout = 15 * time.Millisecond
	w.listIfaces = func() ([]net.Interface, error) { return nil, nil }

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var resyncs atomic.Int32
	done := make(chan struct{})
	go func() {
		defer close(done)
		w.Start(ctx, func() int { return 2 }, func() { resyncs.Add(1) })
	}()

	// Well past the no-peer timeout, but peers exist — no resync expected.
	time.Sleep(100 * time.Millisecond)
	if resyncs.Load() != 0 {
		t.Fatalf("expected no resync with peers present, got %d", resyncs.Load())
	}
	cancel()
	<-done
}

func TestNetworkWatcherStopsOnCancel(t *testing.T) {
	w := NewNetworkWatcher()
	w.pollInterval = 10 * time.Millisecond
	w.noPeerTimeout = 10 * time.Millisecond
	w.listIfaces = func() ([]net.Interface, error) { return nil, nil }

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		defer close(done)
		w.Start(ctx, func() int { return 0 }, func() {})
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("watcher did not stop on cancel")
	}
}
