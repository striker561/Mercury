package sync

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"
)

const (
	// serviceType is the mDNS service type used for Mercury peer discovery.
	serviceType = "_mercury._tcp"

	// domain is the mDNS domain (local link-local).
	domain = "local."

	// resolveTimeout is the maximum time to wait for an mDNS resolution.
	resolveTimeout = 2 * time.Second
)

// hostname returns the machine hostname, falling back to "unknown" on error.
func hostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}

// peerIDFromEntry returns the stable machine ID announced in a service
// entry's TXT record ("id=<uuid>"), or "" for legacy peers without one.
func peerIDFromEntry(entry *zeroconf.ServiceEntry) string {
	for _, kv := range entry.Text {
		if strings.HasPrefix(kv, "id=") {
			return strings.TrimPrefix(kv, "id=")
		}
	}
	return ""
}

// Announce registers this Mercury instance on the LAN via mDNS so other
// instances can discover it.  The server shuts down automatically when
// ctx is cancelled.  deviceID is advertised in the TXT record so peers can
// dedup us across hostname changes (dual-boot).
func Announce(ctx context.Context, port int, instance, deviceID string) error {
	if instance == "" {
		instance = hostname()
	}

	// Advertise our stable machine ID in the TXT record.
	var text []string
	if deviceID != "" {
		text = []string{"id=" + deviceID}
	}

	server, err := zeroconf.Register(
		instance,    // service instance name
		serviceType, // service type: _mercury._tcp
		domain,      // domain: local.
		port,        // port
		text,        // TXT metadata (stable machine id)
		nil,         // all interfaces
	)
	if err != nil {
		return fmt.Errorf("sync mDNS announce: %w", err)
	}

	go func() {
		<-ctx.Done()
		server.Shutdown()
	}()

	return nil
}

// Browse discovers other Mercury instances on the LAN.  It sends discovered
// peers to the added channel.  The goroutine runs until ctx is cancelled.
//
// IMPORTANT: resolver.Browse is non-blocking — it starts background
// goroutines and returns immediately.  We must NOT use a "browse done"
// signal to exit the loop; entries arrive asynchronously on the channel
// until ctx is cancelled.
func Browse(ctx context.Context, added chan<- Peer, deviceID string) error {
	resolver, err := zeroconf.NewResolver()
	if err != nil {
		return fmt.Errorf("sync mDNS resolver: %w", err)
	}

	entries := make(chan *zeroconf.ServiceEntry, 10)

	if err := resolver.Browse(ctx, serviceType, domain, entries); err != nil {
		return fmt.Errorf("sync mDNS browse: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case entry, ok := <-entries:
			if !ok {
				return nil
			}

			// Skip ourselves — by hostname for legacy peers, or by stable
			// machine ID so a dual-boot machine announcing its other OS's
			// hostname doesn't discover itself.
			if entry.Instance == hostname() {
				continue
			}
			if deviceID != "" && peerIDFromEntry(entry) == deviceID {
				continue
			}

			peer := resolveEntry(entry)
			if peer == nil {
				continue
			}

			select {
			case added <- *peer:
			case <-ctx.Done():
				return nil
			}
		}
	}
}

// resolveEntry resolves a zeroconf ServiceEntry into a Peer.  Returns nil
// if the entry cannot be resolved to an IP address.
func resolveEntry(entry *zeroconf.ServiceEntry) *Peer {
	// Resolve the entry if needed — zeroconf.Browse may return entries
	// without IP addresses depending on the network environment.
	if len(entry.AddrIPv4) == 0 && len(entry.AddrIPv6) == 0 {
		resolved, err := resolveInstance(entry.Instance)
		if err != nil || resolved == nil {
			return nil
		}
		entry = resolved
	}

	// Pick the first IPv4 address, falling back to IPv6.
	var addrStr string
	if len(entry.AddrIPv4) > 0 {
		addrStr = net.JoinHostPort(entry.AddrIPv4[0].String(), fmt.Sprintf("%d", entry.Port))
	} else if len(entry.AddrIPv6) > 0 {
		addrStr = net.JoinHostPort(entry.AddrIPv6[0].String(), fmt.Sprintf("%d", entry.Port))
	} else {
		return nil
	}

	// Prefer the stable machine ID announced in TXT (dual-boot dedup);
	// fall back to the instance hostname for legacy peers that don't
	// announce one.  Hostname is kept for display only.
	deviceID := peerIDFromEntry(entry)
	id := deviceID
	if id == "" {
		id = entry.Instance
	}

	return &Peer{
		ID:       id,
		Hostname: entry.Instance,
		Addr:     addrStr,
	}
}

// resolveInstance performs a full mDNS resolution for a specific service
// instance, returning a populated ServiceEntry with IP addresses.
func resolveInstance(instance string) (*zeroconf.ServiceEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), resolveTimeout)
	defer cancel()

	resolver, err := zeroconf.NewResolver()
	if err != nil {
		return nil, fmt.Errorf("sync mDNS resolver: %w", err)
	}

	entries := make(chan *zeroconf.ServiceEntry, 1)
	lookupDone := make(chan struct{})

	go func() {
		defer close(lookupDone)
		// NOTE: do NOT close(entries) — same reason as Browse: the zeroconf
		// library's internal goroutines may still send after Lookup returns.
		if err := resolver.Lookup(ctx, instance, serviceType, domain, entries); err != nil {
			if ctx.Err() == nil {
				log.Printf("[sync] mDNS lookup %s: %v", instance, err)
			}
		}
	}()

	select {
	case entry, ok := <-entries:
		if !ok {
			<-lookupDone
			return nil, fmt.Errorf("no mDNS entry for %s", instance)
		}
		return entry, nil
	case <-ctx.Done():
		<-lookupDone
		return nil, ctx.Err()
	}
}
