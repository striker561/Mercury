package transfer

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"

	syncpkg "mercury/app/backend/sync"
)

// sendFile streams a file to the receiver over a dedicated TCP connection.
func (m *Manager) sendFile(tid, peerAddr, filePath string, fileSize int64) {
	defer m.updateStatus(tid, StatusFailed, 0) // overwritten on success by Done

	f, err := os.Open(filePath)
	if err != nil {
		log.Printf("[transfer] open %s: %v", filePath, err)
		return
	}
	defer f.Close()

	addr := fmt.Sprintf("%s:%d", stripPort(peerAddr), Port)
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		log.Printf("[transfer] dial %s: %v", addr, err)
		return
	}
	defer conn.Close()

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
			hdr := make([]byte, 4)
			binary.BigEndian.PutUint32(hdr, uint32(len(enc)))
			if _, werr := conn.Write(hdr); werr != nil {
				log.Printf("[transfer] write header: %v", werr)
				return
			}
			if _, werr := conn.Write(enc); werr != nil {
				log.Printf("[transfer] write chunk: %v", werr)
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
