package transfer

import (
	"fmt"
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
	var failReason string
	failed := true
	defer func() {
		m.mu.Lock()
		delete(m.cancel, tid)
		m.mu.Unlock()
		if failed {
			if failReason == "" {
				failReason = "transfer interrupted"
			}
			m.fail(tid, failReason)
			log.Printf("[transfer] send %s failed: %s", filepath.Base(filePath), failReason)
		}
	}()

	log.Printf("[transfer] starting send %s to %s (%d bytes)", filepath.Base(filePath), peerAddr, fileSize)
	m.updateStatus(tid, StatusSending, 0)

	cancelCh := make(chan struct{})
	m.mu.Lock()
	m.cancel[tid] = cancelCh
	m.mu.Unlock()

	f, err := os.Open(filePath)
	if err != nil {
		failReason = fileError("open", err)
		return
	}
	defer f.Close()

	addr := peerTCPAddr(peerAddr)
	conn, err := dialPeer(addr, cancelCh)
	if err != nil {
		failReason = fmt.Sprintf("dial %s: %v", addr, err)
		return
	}
	defer conn.Close()

	buf := make([]byte, chunkSize)
	var total int64
	startTime := time.Now()
	progressTick := time.NewTicker(200 * time.Millisecond)
	defer progressTick.Stop()

	for total < fileSize {
		select {
		case <-cancelCh:
			failReason = "cancelled"
			return
		default:
		}

		n, rerr := f.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			wire, perr := packChunkPayload(offerID, chunk)
			if perr != nil {
				failReason = fmt.Sprintf("pack chunk: %v", perr)
				return
			}
			enc, cerr := crypto.Encrypt(wire, m.key)
			if cerr != nil {
				failReason = fmt.Sprintf("encrypt: %v", cerr)
				return
			}

			conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := transport.WriteFrame(conn, transport.MsgFileChunk, enc); err != nil {
				failReason = fmt.Sprintf("send chunk: %v", err)
				return
			}
			total += int64(n)

			select {
			case <-progressTick.C:
				if elapsed := time.Since(startTime).Seconds(); elapsed > 0 {
					m.updateSpeed(tid, int64(float64(total)/elapsed))
				}
				m.updateStatus(tid, StatusSending, total)
			default:
			}
		}
		if rerr != nil {
			if rerr != io.EOF {
				failReason = fmt.Sprintf("read: %v", rerr)
				return
			}
			break
		}
	}

	if total != fileSize {
		failReason = fmt.Sprintf("file changed during send (%d/%d bytes)", total, fileSize)
		return
	}

	endEnc, err := crypto.Encrypt([]byte(offerID), m.key)
	if err != nil {
		failReason = fmt.Sprintf("encrypt end: %v", err)
		return
	}
	conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	if err := transport.WriteFrame(conn, transport.MsgFileEnd, endEnc); err != nil {
		failReason = fmt.Sprintf("send end: %v", err)
		return
	}

	failed = false
	m.updateStatus(tid, StatusDone, fileSize)
	m.updateSpeed(tid, 0)
	log.Printf("[transfer] sent %s (%d bytes)", filepath.Base(filePath), fileSize)
}

func dialPeer(addr string, cancel <-chan struct{}) (net.Conn, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		select {
		case <-cancel:
			return nil, fmt.Errorf("cancelled")
		default:
		}

		dialer := &net.Dialer{Timeout: dialTimeout}
		conn, err := dialer.Dial("tcp", addr)
		if err == nil {
			return conn, nil
		}
		lastErr = err

		if attempt < 2 {
			select {
			case <-time.After(time.Duration(attempt+1) * 500 * time.Millisecond):
			case <-cancel:
				return nil, fmt.Errorf("cancelled")
			}
		}
	}
	return nil, lastErr
}
