package transfer

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

// receiveFile reads decrypted chunks from chunkBuf and writes them to disk.
func (m *Manager) receiveFile(tid string, o *Offer, saveDir string) {
	failed := true
	defer func() {
		if failed {
			m.updateStatus(tid, StatusFailed, 0)
		}
	}()

	log.Printf("[transfer] receiving %s to %s", o.FileName, saveDir)
	m.updateStatus(tid, StatusReceiving, 0)

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
	// Chunks arrive already decrypted via sync → OnFileChunk → chunkBuf.
	// Read until file is complete, with a 30s idle timeout.
	timeout := time.NewTimer(30 * time.Second)
	defer timeout.Stop()

	for received < o.FileSize {
		select {
		case chunk := <-m.chunkBuf:
			if _, err := f.Write(chunk); err != nil {
				log.Printf("[transfer] write: %v", err)
				return
			}
			received += int64(len(chunk))
			timeout.Reset(30 * time.Second) // got data, reset idle timer
		case <-timeout.C:
			log.Printf("[transfer] timeout waiting for chunk (%d/%d)", received, o.FileSize)
			return
		}
	}

	failed = false
	m.updateStatus(tid, StatusDone, received)
	log.Printf("[transfer] received %s (%d bytes)", o.FileName, received)
}
