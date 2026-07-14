package transfer

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	syncpkg "mercury/app/backend/sync"
)

// receiveFile listens for the sender and writes the stream to disk.
func (m *Manager) receiveFile(tid string, o *Offer, saveDir string) {
	defer m.updateStatus(tid, StatusFailed, 0) // overwritten on success

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

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", Port))
	if err != nil {
		log.Printf("[transfer] listen: %v", err)
		return
	}
	defer ln.Close()

	ln.(*net.TCPListener).SetDeadline(time.Now().Add(30 * time.Second))
	conn, err := ln.Accept()
	if err != nil {
		log.Printf("[transfer] accept: %v", err)
		return
	}
	defer conn.Close()

	var received int64
	for received < o.FileSize {
		conn.SetReadDeadline(time.Now().Add(readTimeout))

		hdr := make([]byte, 4)
		if _, err := io.ReadFull(conn, hdr); err != nil {
			log.Printf("[transfer] read header: %v", err)
			return
		}
		plen := binary.BigEndian.Uint32(hdr)
		if plen > chunkSize+1024 {
			log.Printf("[transfer] chunk too large: %d", plen)
			return
		}

		enc := make([]byte, plen)
		if _, err := io.ReadFull(conn, enc); err != nil {
			log.Printf("[transfer] read chunk: %v", err)
			return
		}
		dec, err := syncpkg.Decrypt(enc, m.key)
		if err != nil {
			log.Printf("[transfer] decrypt: %v", err)
			return
		}
		if _, err := f.Write(dec); err != nil {
			log.Printf("[transfer] write: %v", err)
			return
		}
		received += int64(len(dec))
	}

	m.updateStatus(tid, StatusDone, received)
	log.Printf("[transfer] received %s (%d bytes)", o.FileName, received)
}
