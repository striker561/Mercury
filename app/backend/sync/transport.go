package sync

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

const (
	// Port is the single TCP port for all Mercury traffic (clipboard + file).
	Port = 47821

	// Message type byte values.
	MsgClipboard = 0 // encrypted clipboard payload (text/image)
	MsgFileChunk = 1 // encrypted file chunk (no size limit)

	// maxClipboardSize is the max clipboard payload (25 MB). File chunks
	// bypass this limit — they stream on the same port but type 1.
	maxClipboardSize = 25 * 1024 * 1024

	sendTimeout = 5 * time.Second
)

// Wire format for ALL messages on port 47821:
//
//	[1 byte message type][4 bytes big-endian uint32 payload length][payload bytes]
//
// Type 0 (clipboard): payload is encrypted text/image data.
// Type 1 (file chunk): payload is an encrypted 256 KiB file chunk.
//
// The type byte is NOT encrypted — it's only a routing hint. The payload
// is always AES-256-GCM encrypted. Even on a single port, clipboard sync
// and file transfers are independent: the event loop demuxes by type.

// Send dials a peer and sends a clipboard payload (type 0).
func Send(ctx context.Context, addr string, payload []byte) error {
	return SendMsg(ctx, addr, MsgClipboard, payload)
}

// SendMsg dials a peer and sends a payload with the given message type.
// The wire format is [1 byte type][4 bytes length][payload bytes].
func SendMsg(ctx context.Context, addr string, msgType byte, payload []byte) error {
	if msgType == MsgClipboard && len(payload) > maxClipboardSize {
		return fmt.Errorf("sync send: clipboard payload too large (%d > %d)", len(payload), maxClipboardSize)
	}

	dialer := &net.Dialer{Timeout: 2 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("sync send dial %s: %w", addr, err)
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		conn.SetWriteDeadline(deadline)
	} else {
		conn.SetWriteDeadline(time.Now().Add(sendTimeout))
	}

	// Write: [type byte][4-byte length][payload]
	buf := make([]byte, 1+4+len(payload))
	buf[0] = msgType
	binary.BigEndian.PutUint32(buf[1:5], uint32(len(payload)))
	copy(buf[5:], payload)

	if _, err := conn.Write(buf); err != nil {
		return fmt.Errorf("sync send to %s: %w", addr, err)
	}
	return nil
}

// Listen starts a TCP listener on the given port and routes incoming
// messages by type: clipboard payloads go to incoming, file chunks go
// to fileChunks.  The listener runs until ctx is cancelled.
func Listen(ctx context.Context, port int, incoming chan<- []byte, fileChunks chan<- []byte) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("sync listen: %w", err)
	}

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			log.Printf("[sync] accept error: %v", err)
			continue
		}
		go handleConnection(conn, incoming, fileChunks)
	}
}

// handleConnection reads one message from a connection and routes it.
func handleConnection(conn net.Conn, incoming chan<- []byte, fileChunks chan<- []byte) {
	defer conn.Close()

	// Set a read deadline so slow clients don't hold us up.
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	// Read [1 byte type][4 bytes length].
	header := make([]byte, 5)
	if _, err := io.ReadFull(conn, header); err != nil {
		if err != io.EOF {
			log.Printf("[sync] read header: %v", err)
		}
		return
	}

	msgType := header[0]
	payloadLen := binary.BigEndian.Uint32(header[1:5])

	if msgType == MsgClipboard && payloadLen > maxClipboardSize {
		log.Printf("[sync] clipboard payload too large: %d", payloadLen)
		return
	}
	// Safety cap: 10 MB per message (file chunks are 256 KB each,
	// clipboard is <25 MB but we check above for type 0).
	if payloadLen > 10*1024*1024 {
		log.Printf("[sync] payload too large: %d", payloadLen)
		return
	}

	// Read the payload.
	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(conn, payload); err != nil {
		log.Printf("[sync] read payload: %v", err)
		return
	}

	switch msgType {
	case MsgClipboard:
		select {
		case incoming <- payload:
		default:
			log.Printf("[sync] dropped clipboard payload (channel full)")
		}
	case MsgFileChunk:
		select {
		case fileChunks <- payload:
		default:
			log.Printf("[sync] dropped file chunk (channel full)")
		}
	}
}
