package clipboard

import (
	"strings"
	"sync"
)

// Package-level pause state for the clipboard watcher.
var (
	mu           sync.RWMutex
	isPaused     bool
	ignoredProcs []string
)

// PauseCapture pauses clipboard monitoring globally.
func PauseCapture() {
	mu.Lock()
	isPaused = true
	mu.Unlock()
}

// ResumeCapture resumes clipboard monitoring globally.
func ResumeCapture() {
	mu.Lock()
	isPaused = false
	mu.Unlock()
}

// IsPaused reports whether clipboard capture is globally paused.
func IsPaused() bool {
	mu.RLock()
	defer mu.RUnlock()
	return isPaused
}

// SetIgnoredProcesses sets the list of process names to ignore.
func SetIgnoredProcesses(names []string) {
	mu.Lock()
	defer mu.Unlock()
	ignoredProcs = make([]string, len(names))
	for i, n := range names {
		ignoredProcs[i] = strings.ToLower(n)
	}
}
