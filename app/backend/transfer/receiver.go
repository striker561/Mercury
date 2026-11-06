package transfer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// receiveFile writes to a .part file and renames only after the full size.
func (m *Manager) receiveFile(tid, offerID string, o *Offer, saveDir string, stream *receiveStream) {
	dst := uniqueDestination(filepath.Join(saveDir, safeName(o.FileName)))
	part := fmt.Sprintf("%s.part-%s", dst, offerID)

	var failReason string
	success := false

	defer func() {
		m.mu.Lock()
		if m.receiveCh[offerID] == stream {
			delete(m.receiveCh, offerID)
		}
		delete(m.cancel, tid)
		m.mu.Unlock()
		close(stream.done)

		if success {
			m.updateStatus(tid, StatusDone, o.FileSize)
			m.updateSpeed(tid, 0)
			log.Printf("[transfer] received %s (%d bytes)", o.FileName, o.FileSize)
			return
		}
		if failReason == "" {
			failReason = "transfer interrupted"
		}
		m.fail(tid, failReason)
		log.Printf("[transfer] receive %s failed: %s", o.FileName, failReason)
	}()

	log.Printf("[transfer] receiving %s to %s", o.FileName, dst)
	m.updateStatus(tid, StatusReceiving, 0)

	cancelCh := make(chan struct{})
	m.mu.Lock()
	m.cancel[tid] = cancelCh
	m.mu.Unlock()

	if err := os.MkdirAll(saveDir, 0755); err != nil {
		failReason = fileError("create folder", err)
		return
	}

	f, err := os.Create(part)
	if err != nil {
		failReason = fileError("create file", err)
		return
	}
	defer func() {
		f.Close()
		if !success {
			if err := os.Remove(part); err != nil && !os.IsNotExist(err) {
				log.Printf("[transfer] cleanup %s: %v", part, err)
			}
		}
	}()

	var received int64
	startTime := time.Now()
	progressTick := time.NewTicker(200 * time.Millisecond)
	defer progressTick.Stop()
	timeout := time.NewTimer(30 * time.Second)
	defer timeout.Stop()

	writeChunk := func(chunk []byte) bool {
		if _, err := f.Write(chunk); err != nil {
			failReason = fmt.Sprintf("write: %v", err)
			return false
		}
		received += int64(len(chunk))
		timeout.Reset(30 * time.Second)
		select {
		case <-progressTick.C:
			if elapsed := time.Since(startTime).Seconds(); elapsed > 0 {
				m.updateSpeed(tid, int64(float64(received)/elapsed))
			}
			m.updateStatus(tid, StatusReceiving, received)
		default:
		}
		return true
	}

	finish := func() {
		if received != o.FileSize {
			failReason = fmt.Sprintf("connection closed at %d/%d bytes", received, o.FileSize)
			return
		}
		if err := f.Close(); err != nil {
			failReason = fileError("close", err)
			return
		}
		if err := os.Rename(part, dst); err != nil {
			failReason = fileError("rename", err)
			return
		}
		success = true
	}

	for received < o.FileSize {
		select {
		case chunk := <-stream.chunks:
			if !writeChunk(chunk) {
				return
			}
			select {
			case <-cancelCh:
				failReason = "cancelled"
				return
			default:
			}
		case <-stream.end:
			for {
				select {
				case chunk := <-stream.chunks:
					if !writeChunk(chunk) {
						return
					}
				default:
					finish()
					return
				}
			}
		case <-timeout.C:
			failReason = fmt.Sprintf("timed out at %d/%d bytes", received, o.FileSize)
			return
		case <-cancelCh:
			failReason = "cancelled"
			return
		}
	}

	finish()
}

func safeName(name string) string {
	name = filepath.Base(strings.ReplaceAll(name, `\`, "/"))
	if name == "" || name == "." || name == ".." {
		return "download"
	}
	return name
}

func uniqueDestination(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	for i := 1; i < 10000; i++ {
		candidate := fmt.Sprintf("%s (%d)%s", base, i, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
	return path
}
