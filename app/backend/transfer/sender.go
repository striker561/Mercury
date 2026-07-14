package transfer

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"

	syncpkg "mercury/app/backend/sync"
)

// sendFile streams a file by sending each encrypted chunk as a MsgFileChunk
// message over the sync port.  The receiver's sync listener demuxes file
// chunks to the transfer manager.
func (m *Manager) sendFile(tid, peerAddr, filePath string, fileSize int64) {
	defer m.updateStatus(tid, StatusFailed, 0)

	f, err := os.Open(filePath)
	if err != nil {
		log.Printf("[transfer] open %s: %v", filePath, err)
		return
	}
	defer f.Close()

	addr := stripPort(peerAddr) + ":47821"
	buf := make([]byte, chunkSize)
	var total int64

	for total < fileSize {
		n, rerr := f.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			enc, cerr := syncpkg.Encrypt(chunk, m.key)
			if cerr != nil {
				log.Printf("[transfer] encrypt: %v", cerr)
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
			err := syncpkg.SendMsg(ctx, addr, syncpkg.MsgFileChunk, enc)
			cancel()
			if err != nil {
				log.Printf("[transfer] send chunk: %v", err)
				return
			}
			total += int64(n)
		}
		if rerr != nil {
			if rerr != io.EOF {
				log.Printf("[transfer] read: %v", rerr)
				return
			}
			break
		}
	}

	m.updateStatus(tid, StatusDone, fileSize)
	log.Printf("[transfer] sent %s (%d bytes)", filepath.Base(filePath), fileSize)
}
