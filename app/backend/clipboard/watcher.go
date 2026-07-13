// Package clipboard watches the OS clipboard for text and image changes,
// emitting them with a 150ms debounce. Uses golang.design/x/clipboard
// for cross-platform clipboard access.
package clipboard

import (
	"context"
	"log"
	"sync"
	"time"

	"golang.design/x/clipboard"
)

// ChangeType indicates what kind of clipboard content changed.
type ChangeType int

const (
	ChangeText ChangeType = iota
	ChangeImage
)

// Change represents a clipboard content change.
type Change struct {
	Type  ChangeType
	Text  string // populated for ChangeText
	Image []byte // populated for ChangeImage (PNG bytes)
}

// Reader abstracts clipboard reads for testability.
type Reader interface {
	ReadText() string
	ReadImage() []byte
}

// liveReader uses the real golang.design/x/clipboard package.
type liveReader struct{}

func (liveReader) ReadText() string {
	if b := clipboard.Read(clipboard.FmtText); b != nil {
		return string(b)
	}
	return ""
}

func (liveReader) ReadImage() []byte {
	return clipboard.Read(clipboard.FmtImage)
}

// Watcher monitors the OS clipboard and emits changes.
type Watcher struct {
	reader       Reader
	prevText     string
	prevImage    []byte
	mu           sync.Mutex
	paused       bool
	onChange     func(Change)
	lastEvent    time.Time
	pollInterval time.Duration
	debounceDur  time.Duration
}

// NewWatcher creates a new clipboard watcher using the real clipboard.
func NewWatcher() *Watcher {
	return &Watcher{
		reader:       liveReader{},
		pollInterval: 150 * time.Millisecond,
		debounceDur:  150 * time.Millisecond,
	}
}

// newWatcherWithReader creates a watcher with a custom reader (for testing).
func newWatcherWithReader(r Reader) *Watcher {
	return &Watcher{
		reader:       r,
		pollInterval: 150 * time.Millisecond,
		debounceDur:  150 * time.Millisecond,
	}
}

// OnChange registers a callback for clipboard changes. Must be called before Start.
func (w *Watcher) OnChange(cb func(Change)) {
	w.onChange = cb
}

// Start begins polling the clipboard. Blocks until ctx is cancelled.
// Callbacks fire synchronously within this goroutine.
func (w *Watcher) Start(ctx context.Context) {
	if w.onChange == nil {
		log.Println("[clipboard] no callback registered, watcher idle")
		return
	}

	w.mu.Lock()
	w.prevText = w.reader.ReadText()
	w.prevImage = w.reader.ReadImage()
	w.lastEvent = time.Now()
	w.mu.Unlock()

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.poll()
		}
	}
}

// poll checks the clipboard for changes with debounce.
// Callbacks fire synchronously — the caller's goroutine owns the timing.
func (w *Watcher) poll() {
	w.mu.Lock()

	if w.paused {
		w.mu.Unlock()
		return
	}

	now := time.Now()
	if now.Sub(w.lastEvent) < w.debounceDur {
		w.mu.Unlock()
		return
	}

	text := w.reader.ReadText()
	img := w.reader.ReadImage()

	if text != w.prevText {
		w.prevText = text
		w.lastEvent = now
		cb := w.onChange
		w.mu.Unlock()
		if cb != nil {
			cb(Change{Type: ChangeText, Text: text})
		}
		return
	}

	if img != nil && !bytesEqual(img, w.prevImage) {
		w.prevImage = make([]byte, len(img))
		copy(w.prevImage, img)
		w.lastEvent = now
		cb := w.onChange
		w.mu.Unlock()
		if cb != nil {
			cb(Change{Type: ChangeImage, Image: img})
		}
		return
	}

	w.mu.Unlock()
}

// Pause temporarily stops clipboard monitoring.
func (w *Watcher) Pause() {
	w.mu.Lock()
	w.paused = true
	w.mu.Unlock()
}

// Resume restarts clipboard monitoring after a pause.
func (w *Watcher) Resume() {
	w.mu.Lock()
	w.paused = false
	w.mu.Unlock()
}

// IsPaused reports whether the watcher is paused.
func (w *Watcher) IsPaused() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.paused
}

// bytesEqual compares two byte slices for equality.
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
