package transfer

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"mercury/app/backend/crypto"
	"mercury/app/backend/transport"
)

// sendFile streams a file over the sync port as MsgFileChunk messages.
func (m *Manager) sendFile(tid, peerAddr, filePath string, fileSize int64) {
	failed := true
	defer func() {
		if failed {
			m.updateStatus(tid, StatusFailed, 0)
		}
	}()

	log.Printf("[transfer] starting send %s to %s (%d bytes)", filepath.Base(filePath), peerAddr, fileSize)
	m.updateStatus(tid, StatusSending, 0)

	f, err := os.Open(filePath)
	if err != nil {
		log.Printf("[transfer] open %s: %v", filePath, err)
		return
	}
	defer f.Close()

	addr := fmt.Sprintf("%s:%d", stripPort(peerAddr), transport.Port)
	buf := make([]byte, chunkSize)
	var total int64

	// Progress ticker — update UI every 200ms.
	progressTick := time.NewTicker(200 * time.Millisecond)
	defer progressTick.Stop()

	// Read → encrypt → send loop.  One TCP conn per chunk (LAN; overhead is fine).
	for total < fileSize {
		n, rerr := f.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			enc, cerr := crypto.Encrypt(chunk, m.key)
			if cerr != nil {
				log.Printf("[transfer] encrypt: %v", cerr)
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
			err := transport.SendMsg(ctx, addr, transport.MsgFileChunk, enc)
			cancel()
			if err != nil {
				log.Printf("[transfer] send chunk: %v", err)
				return
			}
			total += int64(n)

			// Push progress to frontend periodically.
			select {
			case <-progressTick.C:
				m.updateStatus(tid, StatusSending, total)
			default:
			}
		}
		if rerr != nil {
			if rerr != io.EOF {
				log.Printf("[transfer] read: %v", rerr)
				return
			}
			break
		}
	}

	failed = false
	m.updateStatus(tid, StatusDone, fileSize)
	log.Printf("[transfer] sent %s (%d bytes)", filepath.Base(filePath), fileSize)
}
