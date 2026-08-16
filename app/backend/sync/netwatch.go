package sync

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

// NetworkWatcher monitors the local network and triggers discovery resyncs.
//
// Why this exists: mDNS discovery (grandcat/zeroconf) snapshots the multicast
// interfaces once when the resolver is created.  If the network isn't ready
// when the app first launches — common right after login — the resolver never
// joins the mDNS group, so no peers are found until discovery is restarted.
// Watching for interface changes (Wi-Fi drops, VPN, Ethernet re-plug) and for
// "no peers for a while" lets us restart discovery automatically, which is
// exactly what the manual "toggle sync off/on" used to do.
type NetworkWatcher struct {
	// pollInterval is how often the network is re-checked.
	pollInterval time.Duration
	// noPeerTimeout is how long with zero peers before a retry is triggered.
	noPeerTimeout time.Duration
	// listIfaces is injectable for tests.
	listIfaces func() ([]net.Interface, error)
}

// NewNetworkWatcher creates a watcher with sensible defaults.
func NewNetworkWatcher() *NetworkWatcher {
	return &NetworkWatcher{
		pollInterval:  3 * time.Second,
		noPeerTimeout: 15 * time.Second,
		listIfaces:    net.Interfaces,
	}
}

// interfaceFingerprint returns a string that changes when the set of
// up-and-running interfaces or their addresses change.  Only interfaces that
// are up are considered — a suspended Wi-Fi shouldn't trigger anything; we
// react when the network comes back (or changes) and peers can't be found.
func (w *NetworkWatcher) interfaceFingerprint() string {
	ifaces, err := w.listIfaces()
	if err != nil {
		return ""
	}
	var b strings.Builder
	for _, ifc := range ifaces {
		if ifc.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, _ := ifc.Addrs()
		fmt.Fprintf(&b, "%s|%d|", ifc.Name, ifc.MTU)
		for _, a := range addrs {
			fmt.Fprintf(&b, "%s,", a.String())
		}
		b.WriteString(";")
	}
	return b.String()
}

// Start runs the watch loop until ctx is cancelled.  peerCount reports the
// current number of known peers; resync is called when the network fingerprint
// changes or when there have been no peers for noPeerTimeout (a first-launch
// retry).
func (w *NetworkWatcher) Start(ctx context.Context, peerCount func() int, resync func()) {
	last := w.interfaceFingerprint()
	var emptySince time.Time

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cur := w.interfaceFingerprint()
			if cur != last {
				last = cur
				emptySince = time.Time{}
				resync()
				continue
			}

			if peerCount() == 0 {
				if emptySince.IsZero() {
					emptySince = time.Now()
				} else if time.Since(emptySince) >= w.noPeerTimeout {
					// Still nothing after the timeout — restart discovery.
					// Reset the timer so a persistent failure keeps retrying
					// at the timeout cadence instead of spinning hot.
					emptySince = time.Now()
					resync()
				}
			} else {
				emptySince = time.Time{}
			}
		}
	}
}
