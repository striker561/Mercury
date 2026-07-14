package transfer

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

// receiveFile reads decrypted chunks from the per-offer channel and writes them to disk.
func (m *Manager) receiveFile(tid, offerID string, o *Offer, saveDir string, ch <-chan []byte) {
	failed := true
	defer func() {
		m.mu.Lock()
		delete(m.receiveCh, offerID)
		delete(m.cancel, tid)
		m.mu.Unlock()
		if failed {
			m.updateStatus(tid, StatusFailed, 0)
			// Clean up the partial file on failure/cancel.
			dst := filepath.Join(saveDir, o.FileName)
			if err := os.Remove(dst); err != nil && !os.IsNotExist(err) {
				log.Printf("[transfer] cleanup %s: %v", dst, err)
			}
		}
	}()

	log.Printf("[transfer] receiving %s to %s", o.FileName, saveDir)
	m.updateStatus(tid, StatusReceiving, 0)

	// Register cancel channel.
	cancelCh := make(chan struct{})
	m.mu.Lock()
	m.cancel[tid] = cancelCh
	m.mu.Unlock()

	if err := os.MkdirAll(saveDir, 0755); err != nil {
		log.Printf("[transfer] mkdir %s: %v", saveDir, err)
		return
	}

	dst := filepath.Join(saveDir, o.FileName)
	f, err := os.Create(dst)
	if err != nil {
		log.Printf("[transfer] create %s: %v", dst, err)
		return
	}
	defer f.Close()

	var received int64
	startTime := time.Now()

	// Progress ticker — update UI every 200ms.
	progressTick := time.NewTicker(200 * time.Millisecond)
	defer progressTick.Stop()

	// Chunks arrive already decrypted via sync → OnWireMessage → per-offer channel.
	// Read until file is complete or the sender closes the channel (MsgFileEnd).
	timeout := time.NewTimer(30 * time.Second)
	defer timeout.Stop()

	for received < o.FileSize {
		select {
		case chunk, ok := <-ch:
			if !ok {
				if received != o.FileSize {
					log.Printf("[transfer] connection closed early (%d/%d)", received, o.FileSize)
					return
				}
				failed = false
				m.updateStatus(tid, StatusDone, received)
				m.updateSpeed(tid, 0)
				log.Printf("[transfer] received %s (%d bytes)", o.FileName, received)
				return
			}

			if _, err := f.Write(chunk); err != nil {
				log.Printf("[transfer] write: %v", err)
				return
			}
			received += int64(len(chunk))
			timeout.Reset(30 * time.Second)

			// Push progress to frontend periodically with speed.
			select {
			case <-progressTick.C:
				elapsed := time.Since(startTime).Seconds()
				if elapsed > 0 {
					m.updateSpeed(tid, int64(float64(received)/elapsed))
				}
				m.updateStatus(tid, StatusReceiving, received)
			default:
			}

			// Check for cancel.
			select {
			case <-cancelCh:
				log.Printf("[transfer] receive %s cancelled at %d/%d", o.FileName, received, o.FileSize)
				return
			default:
			}

		case <-timeout.C:
			log.Printf("[transfer] timeout waiting for chunk (%d/%d)", received, o.FileSize)
			return
		}
	}

	failed = false
	m.updateStatus(tid, StatusDone, received)
	m.updateSpeed(tid, 0)
	log.Printf("[transfer] received %s (%d bytes)", o.FileName, received)
}
