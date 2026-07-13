// Package services provides the business logic layer between the Wails frontend
// and the backend sync/clipboard engines.
package services

import (
	"context"

	"mercury/app/backend/clipboard"
)

// ClipboardService wraps the clipboard watcher and OS clipboard writes.
// Uses golang.design/x/clipboard internally — no OS-specific files needed.
type ClipboardService struct {
	watcher   *clipboard.Watcher
	stopWatch context.CancelFunc
}

// NewClipboardService creates a new clipboard service.
func NewClipboardService() *ClipboardService {
	return &ClipboardService{}
}

// Start begins watching the clipboard for changes.
// The onChange callback fires when text or image content changes.
func (s *ClipboardService) Start(onChange func(clipboard.Change)) {
	ctx, cancel := context.WithCancel(context.Background())
	s.stopWatch = cancel
	s.watcher = clipboard.NewWatcher()
	s.watcher.OnChange(onChange)
	go s.watcher.Start(ctx)
}

// Stop halts clipboard monitoring.
func (s *ClipboardService) Stop() {
	if s.stopWatch != nil {
		s.stopWatch()
	}
}

// Pause temporarily stops clipboard monitoring.
func (s *ClipboardService) Pause() {
	if s.watcher != nil {
		s.watcher.Pause()
	}
}

// Resume restarts clipboard monitoring.
func (s *ClipboardService) Resume() {
	if s.watcher != nil {
		s.watcher.Resume()
	}
}
