package sync

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
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
// instances can discover it.  deviceID is advertised in the TXT record so
// peers can dedup us across hostname changes (dual-boot).  The returned stop
// function shuts the registration down and is safe to call more than once.
func Announce(ctx context.Context, port int, instance, deviceID string) (func(), error) {
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
		return nil, fmt.Errorf("sync mDNS announce: %w", err)
	}

	var once sync.Once
	stop := func() {
		once.Do(server.Shutdown)
	}

	go func() {
		<-ctx.Done()
		stop()
	}()

	return stop, nil
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

	addrs := entryAddrs(entry)
	if len(addrs) == 0 {
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
		Addr:     addrs[0],
		Addrs:    addrs,
	}
}

// entryAddrs returns every usable address, best candidate first.  A
// multi-homed machine advertises addresses for VPN, bridge and LAN
// interfaces; the first A record is not necessarily reachable.
func entryAddrs(entry *zeroconf.ServiceEntry) []string {
	var raw []string
	for _, ip := range entry.AddrIPv4 {
		raw = append(raw, net.JoinHostPort(ip.String(), strconv.Itoa(entry.Port)))
	}
	for _, ip := range entry.AddrIPv6 {
		raw = append(raw, net.JoinHostPort(ip.String(), strconv.Itoa(entry.Port)))
	}
	return rankAddrs(raw)
}

func rankAddrs(addrs []string) []string {
	locals := localNetworks()
	type scored struct {
		addr  string
		score int
	}
	out := make([]scored, 0, len(addrs))
	seen := make(map[string]bool, len(addrs))
	for _, a := range addrs {
		if seen[a] {
			continue
		}
		seen[a] = true
		s := addrScore(a, locals)
		if s < 0 {
			continue
		}
		out = append(out, scored{addr: a, score: s})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].score > out[j].score })

	ranked := make([]string, len(out))
	for i := range out {
		ranked[i] = out[i].addr
	}
	return ranked
}

// addrScore ranks same-subnet addresses above other private or public ones,
// and rejects addresses that can never reach a peer.
func addrScore(addr string, locals []*net.IPNet) int {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return -1
	}
	ip := net.ParseIP(host)
	if ip == nil || ip.IsUnspecified() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return -1
	}
	for _, n := range locals {
		if n.Contains(ip) {
			return 2
		}
	}
	if ip.IsPrivate() {
		return 1
	}
	return 0
}

func localNetworks() []*net.IPNet {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var out []*net.IPNet
	for _, ifc := range ifaces {
		if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := ifc.Addrs()
		for _, a := range addrs {
			n, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			if n.IP.IsLoopback() || n.IP.IsLinkLocalUnicast() {
				continue
			}
			out = append(out, n)
		}
	}
	return out
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
