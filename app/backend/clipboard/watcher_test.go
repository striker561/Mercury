package clipboard

import (
	"context"
	"testing"
	"time"
)

type mockReader struct {
	text  string
	image []byte
}

func (m *mockReader) ReadText() string  { return m.text }
func (m *mockReader) ReadImage() []byte { return m.image }

func TestBytesEqual(t *testing.T) {
	cases := []struct {
		name string
		a, b []byte
		want bool
	}{
		{"both nil", nil, nil, true},
		{"both empty", []byte{}, []byte{}, true},
		{"equal", []byte("hello"), []byte("hello"), true},
		{"different", []byte("hello"), []byte("world"), false},
		{"diff length", []byte("hi"), []byte("hello"), false},
		{"one nil", nil, []byte("a"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := bytesEqual(tc.a, tc.b); got != tc.want {
				t.Errorf("bytesEqual(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestPauseResume(t *testing.T) {
	w := newWatcherWithReader(&mockReader{})
	if w.IsPaused() {
		t.Fatal("new watcher not paused")
	}
	w.Pause()
	if !w.IsPaused() {
		t.Fatal("should be paused")
	}
	w.Resume()
	if w.IsPaused() {
		t.Fatal("should not be paused after resume")
	}
}

func TestGlobalPauseResume(t *testing.T) {
	ResumeCapture()
	if IsPaused() {
		t.Fatal("not paused initially")
	}
	PauseCapture()
	if !IsPaused() {
		t.Fatal("should be paused")
	}
	ResumeCapture()
	if IsPaused() {
		t.Fatal("should not be paused")
	}
}

func TestPollDetectsTextChange(t *testing.T) {
	mock := &mockReader{text: "old"}
	w := newWatcherWithReader(mock)
	w.debounceDur = 0 // disable debounce for deterministic testing

	var got Change
	w.OnChange(func(c Change) { got = c })

	// First poll initializes state
	w.poll()
	if got.Text != "old" {
		t.Fatalf("expected init 'old', got %q", got.Text)
	}

	// Change text and poll again
	mock.text = "new value"
	w.poll()
	if got.Type != ChangeText {
		t.Errorf("expected ChangeText, got %v", got.Type)
	}
	if got.Text != "new value" {
		t.Errorf("expected 'new value', got %q", got.Text)
	}

	// Same content — no callback
	got = Change{}
	w.poll()
	if got.Text != "" {
		t.Fatal("expected no callback for same content")
	}
}

func TestPollDetectsImageChange(t *testing.T) {
	mock := &mockReader{image: []byte{1, 2, 3}}
	w := newWatcherWithReader(mock)
	w.debounceDur = 0

	var got Change
	w.OnChange(func(c Change) { got = c })

	w.poll() // init
	mock.image = []byte{4, 5, 6}
	w.poll()

	if got.Type != ChangeImage {
		t.Errorf("expected ChangeImage, got %v", got.Type)
	}
}

func TestDebounce(t *testing.T) {
	mock := &mockReader{text: "v1"}
	w := newWatcherWithReader(mock)
	w.debounceDur = time.Hour // effectively infinite debounce

	count := 0
	w.OnChange(func(c Change) { count++ })

	w.poll() // init fires despite debounce (lastEvent = now, but debounceDur is huge)
	// Actually, debounce will block this... let me reset after init
	count = 0

	mock.text = "v2"
	// lastEvent was set by init, so debounce will block
	w.poll()
	if count != 0 {
		t.Fatal("expected 0 callbacks with long debounce")
	}
}

func TestDebounceRespectsInterval(t *testing.T) {
	mock := &mockReader{text: "a"}
	w := newWatcherWithReader(mock)
	w.debounceDur = 50 * time.Millisecond

	count := 0
	w.OnChange(func(c Change) { count++ })
	w.poll() // init
	mock.text = "b"

	// Too fast — debounce blocks
	w.poll()
	if count != 1 {
		t.Fatalf("expected 1 (init), got %d", count)
	}

	// Wait past debounce
	w.mu.Lock()
	w.lastEvent = time.Now().Add(-100 * time.Millisecond)
	w.mu.Unlock()

	w.poll()
	if count != 2 {
		t.Fatalf("expected 2 after waiting, got %d", count)
	}
}

func TestPausedWatcherSkipsPoll(t *testing.T) {
	mock := &mockReader{text: "initial"}
	w := newWatcherWithReader(mock)
	w.debounceDur = 0

	count := 0
	w.OnChange(func(c Change) { count++ })

	w.poll() // init
	count = 0

	w.Pause()
	mock.text = "changed"
	w.poll()
	if count != 0 {
		t.Fatal("should not fire while paused")
	}

	w.Resume()
	w.poll()
	if count != 1 {
		t.Fatalf("expected 1 after resume, got %d", count)
	}
}

func TestWatcherStopsOnCancel(t *testing.T) {
	w := newWatcherWithReader(&mockReader{})
	w.OnChange(func(c Change) {})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		w.Start(ctx)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("watcher did not stop on cancel")
	}
}
