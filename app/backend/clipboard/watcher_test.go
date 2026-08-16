package clipboard

import (
	"context"
	"testing"
	"time"
)

type mockReader struct {
	text  string
	image []byte
	files []string
}

func (m *mockReader) ReadText() string  { return m.text }
func (m *mockReader) ReadImage() []byte { return m.image }
func (m *mockReader) ReadFiles() []string {
	return m.files
}

// changeRecorder collects clipboard-change callbacks on a buffered channel so
// tests can wait deterministically.  The watcher dispatches callbacks
// asynchronously (`go onChange(...)`), so reading a shared variable right after
// poll() races and flakes under `go test ./...`.
type changeRecorder struct {
	ch chan Change
}

func newChangeRecorder() *changeRecorder {
	return &changeRecorder{ch: make(chan Change, 8)}
}

func (r *changeRecorder) on() func(Change) {
	return func(c Change) { r.ch <- c }
}

// next waits for the next change, failing the test if none arrives in time.
func (r *changeRecorder) next(t *testing.T) Change {
	t.Helper()
	select {
	case c := <-r.ch:
		return c
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for clipboard change")
		return Change{}
	}
}

// empty reports whether no change is pending, without blocking.
func (r *changeRecorder) empty() bool {
	select {
	case <-r.ch:
		return false
	default:
		return true
	}
}

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

	rec := newChangeRecorder()
	w.OnChange(rec.on())

	// First poll initializes state and fires for the initial value.
	w.poll()
	if got := rec.next(t); got.Text != "old" {
		t.Fatalf("expected init 'old', got %q", got.Text)
	}

	// Change text and poll again.
	mock.text = "new value"
	w.poll()
	got := rec.next(t)
	if got.Type != ChangeText {
		t.Errorf("expected ChangeText, got %v", got.Type)
	}
	if got.Text != "new value" {
		t.Errorf("expected 'new value', got %q", got.Text)
	}

	// Same content — no callback.
	w.poll()
	if !rec.empty() {
		t.Fatal("expected no callback for same content")
	}
}

func TestPollDetectsImageChange(t *testing.T) {
	mock := &mockReader{image: []byte{1, 2, 3}}
	w := newWatcherWithReader(mock)
	w.debounceDur = 0

	rec := newChangeRecorder()
	w.OnChange(rec.on())

	w.poll() // init — fires for {1,2,3}
	rec.next(t)

	mock.image = []byte{4, 5, 6}
	w.poll()

	if got := rec.next(t); got.Type != ChangeImage {
		t.Errorf("expected ChangeImage, got %v", got.Type)
	}
}

func TestDebounce(t *testing.T) {
	mock := &mockReader{text: "v1"}
	w := newWatcherWithReader(mock)
	w.debounceDur = time.Hour // effectively infinite debounce

	rec := newChangeRecorder()
	w.OnChange(rec.on())

	w.poll() // init fires despite the debounce (lastEvent is the zero time)
	rec.next(t)

	mock.text = "v2"
	// lastEvent was set by init, so the huge debounce blocks this poll.
	w.poll()
	if !rec.empty() {
		t.Fatal("expected 0 callbacks with long debounce")
	}
}

func TestDebounceRespectsInterval(t *testing.T) {
	mock := &mockReader{text: "a"}
	w := newWatcherWithReader(mock)
	w.debounceDur = 50 * time.Millisecond

	rec := newChangeRecorder()
	w.OnChange(rec.on())
	w.poll() // init fires
	rec.next(t)

	mock.text = "b"

	// Too fast — debounce blocks.
	w.poll()
	if !rec.empty() {
		t.Fatal("expected no callback while inside the debounce window")
	}

	// Wait past debounce.
	w.mu.Lock()
	w.lastEvent = time.Now().Add(-100 * time.Millisecond)
	w.mu.Unlock()

	w.poll()
	rec.next(t) // fires for "b"
}

func TestPausedWatcherSkipsPoll(t *testing.T) {
	mock := &mockReader{text: "initial"}
	w := newWatcherWithReader(mock)
	w.debounceDur = 0

	rec := newChangeRecorder()
	w.OnChange(rec.on())

	w.poll() // init fires
	rec.next(t)

	w.Pause()
	mock.text = "changed"
	w.poll()
	if !rec.empty() {
		t.Fatal("should not fire while paused")
	}

	w.Resume()
	w.poll()
	rec.next(t) // fires for "changed"
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

func TestPollDetectsFileChange(t *testing.T) {
	mock := &mockReader{}
	w := newWatcherWithReader(mock)
	w.debounceDur = 0 // disable debounce for deterministic testing

	rec := newChangeRecorder()
	w.OnChange(rec.on())

	w.poll() // init — no files on clipboard yet

	mock.files = []string{`C:\Users\me\Desktop\photo.png`}
	w.poll()
	got := rec.next(t)
	if got.Type != ChangeText {
		t.Fatalf("expected ChangeText, got %v", got.Type)
	}
	if got.Text != `C:\Users\me\Desktop\photo.png` {
		t.Fatalf("expected file path, got %q", got.Text)
	}
}

func TestPollFileTakesFirstPath(t *testing.T) {
	mock := &mockReader{}
	w := newWatcherWithReader(mock)
	w.debounceDur = 0

	rec := newChangeRecorder()
	w.OnChange(rec.on())

	w.poll() // init
	mock.files = []string{`/tmp/a.txt`, `/tmp/b.txt`}
	w.poll()

	if got := rec.next(t); got.Text != `/tmp/a.txt` {
		t.Fatalf("expected first file path, got %q", got.Text)
	}
}

func TestPollSameFileDoesNotRefire(t *testing.T) {
	mock := &mockReader{}
	w := newWatcherWithReader(mock)
	w.debounceDur = 0

	rec := newChangeRecorder()
	w.OnChange(rec.on())

	mock.files = []string{`/tmp/a.txt`}
	w.poll() // fires once
	rec.next(t)
	w.poll() // same path still on clipboard — must not refire

	if !rec.empty() {
		t.Fatal("expected no callback for the same file path")
	}
}

func TestPollFileToTextTransition(t *testing.T) {
	mock := &mockReader{}
	w := newWatcherWithReader(mock)
	w.debounceDur = 0

	rec := newChangeRecorder()
	w.OnChange(rec.on())

	w.poll() // init
	mock.files = []string{`/tmp/a.txt`}
	w.poll()
	if got := rec.next(t); got.Type != ChangeText || got.Text != `/tmp/a.txt` {
		t.Fatalf("expected file path, got %+v", got)
	}

	// File copy replaced by plain text.
	mock.files = nil
	mock.text = "hello"
	w.poll()
	if got := rec.next(t); got.Type != ChangeText || got.Text != "hello" {
		t.Fatalf("expected text transition, got %+v", got)
	}
}
