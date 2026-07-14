package transfer

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

// receiveFile reads decrypted chunks from chunkBuf and writes them to disk.
// The sync manager demuxes MsgFileChunk messages and decrypts them before
// pushing onto chunkBuf — no separate listener needed.
func (m *Manager) receiveFile(tid string, o *Offer, saveDir string) {
	defer m.updateStatus(tid, StatusFailed, 0)

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
	// Chunks arrive already decrypted via the sync manager's OnFileChunk
	// callback.  We read from chunkBuf until the file is complete.
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
			// Reset the timeout on each successful chunk.
			timeout.Reset(30 * time.Second)
		case <-timeout.C:
			log.Printf("[transfer] timeout waiting for chunk (%d/%d)", received, o.FileSize)
			return
		}
	}

	m.updateStatus(tid, StatusDone, received)
	log.Printf("[transfer] received %s (%d bytes)", o.FileName, received)
}
