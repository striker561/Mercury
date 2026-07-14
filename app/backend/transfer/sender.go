package transfer

import (
	"context"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"mercury/app/backend/crypto"
	"mercury/app/backend/transport"
)

// sendFile streams a file over a single persistent TCP connection,
// sending each chunk as a separate frame on the same connection.
func (m *Manager) sendFile(tid, offerID, peerAddr, filePath string, fileSize int64) {
	failed := true
	defer func() {
		m.mu.Lock()
		delete(m.cancel, tid)
		m.mu.Unlock()
		if failed {
			m.updateStatus(tid, StatusFailed, 0)
		}
	}()

	log.Printf("[transfer] starting send %s to %s (%d bytes)", filepath.Base(filePath), peerAddr, fileSize)
	m.updateStatus(tid, StatusSending, 0)

	// Register cancel channel.
	cancelCh := make(chan struct{})
	m.mu.Lock()
	m.cancel[tid] = cancelCh
	m.mu.Unlock()

	f, err := os.Open(filePath)
	if err != nil {
		log.Printf("[transfer] open %s: %v", filePath, err)
		return
	}
	defer f.Close()

	addr := peerTCPAddr(peerAddr)

	// Open a single persistent TCP connection for the whole transfer.
	dialer := &net.Dialer{Timeout: dialTimeout}
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	cancel()
	if err != nil {
		log.Printf("[transfer] dial %s: %v", addr, err)
		return
	}
	defer conn.Close()

	buf := make([]byte, chunkSize)
	var total int64
	startTime := time.Now()

	// Progress ticker — update UI every 200ms.
	progressTick := time.NewTicker(200 * time.Millisecond)
	defer progressTick.Stop()

	// Read → encrypt → send loop over a single connection.
	// The receiver side reads frames in a loop until the connection closes.
	for total < fileSize {
		n, rerr := f.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			wire, perr := packChunkPayload(offerID, chunk)
			if perr != nil {
				log.Printf("[transfer] pack chunk: %v", perr)
				return
			}
			enc, cerr := crypto.Encrypt(wire, m.key)
			if cerr != nil {
				log.Printf("[transfer] encrypt: %v", cerr)
				return
			}

			conn.SetWriteDeadline(time.Now().Add(dialTimeout))
			if err := transport.WriteFrame(conn, transport.MsgFileChunk, enc); err != nil {
				log.Printf("[transfer] send chunk: %v", err)
				return
			}
			total += int64(n)

			// Push progress to frontend periodically with speed.
			select {
			case <-progressTick.C:
				elapsed := time.Since(startTime).Seconds()
				if elapsed > 0 {
					m.updateSpeed(tid, int64(float64(total)/elapsed))
				}
				m.updateStatus(tid, StatusSending, total)
			default:
			}

			// Check for cancel.
			select {
			case <-cancelCh:
				log.Printf("[transfer] send %s cancelled at %d/%d", filepath.Base(filePath), total, fileSize)
				return
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

	endEnc, err := crypto.Encrypt([]byte(offerID), m.key)
	if err != nil {
		log.Printf("[transfer] encrypt end: %v", err)
		return
	}
	conn.SetWriteDeadline(time.Now().Add(dialTimeout))
	if err := transport.WriteFrame(conn, transport.MsgFileEnd, endEnc); err != nil {
		log.Printf("[transfer] send end: %v", err)
		return
	}

	failed = false
	m.updateStatus(tid, StatusDone, fileSize)
	m.updateSpeed(tid, 0)
	log.Printf("[transfer] sent %s (%d bytes)", filepath.Base(filePath), fileSize)
}
