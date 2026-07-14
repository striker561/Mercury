// Package transport provides the TCP wire protocol shared by clipboard sync
// and file transfer.  Wire format:
//
//	[1 byte message type][4 bytes big-endian uint32 payload length][payload bytes]
//
// The type byte is a routing hint (not encrypted).  Payloads are always
// AES-256-GCM encrypted by the caller.
package transport

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
	// Port is the single TCP port for all Mercury traffic.
	Port = 47821

	MsgClipboard = 0 // encrypted clipboard payload
	MsgFileChunk = 1 // encrypted file chunk

	// MaxClipboardSize is the max clipboard payload (25 MB). File chunks bypass this.
	MaxClipboardSize = 25 * 1024 * 1024
	// SendTimeout is the timeout for a single send operation.
	SendTimeout = 5 * time.Second
)

// OnMessage is called for every received message.
type OnMessage func(msgType byte, payload []byte)

// Send sends a clipboard payload (MsgClipboard).
func Send(ctx context.Context, addr string, payload []byte) error {
	return SendMsg(ctx, addr, MsgClipboard, payload)
}

// SendMsg sends a payload with the given message type.
func SendMsg(ctx context.Context, addr string, msgType byte, payload []byte) error {
	if msgType == MsgClipboard && len(payload) > MaxClipboardSize {
		return fmt.Errorf("send: clipboard payload too large (%d > %d)", len(payload), MaxClipboardSize)
	}

	dialer := &net.Dialer{Timeout: 2 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("send dial %s: %w", addr, err)
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		conn.SetWriteDeadline(deadline)
	} else {
		conn.SetWriteDeadline(time.Now().Add(SendTimeout))
	}

	buf := make([]byte, 1+4+len(payload))
	buf[0] = msgType
	binary.BigEndian.PutUint32(buf[1:5], uint32(len(payload)))
	copy(buf[5:], payload)

	if _, err := conn.Write(buf); err != nil {
		return fmt.Errorf("send to %s: %w", addr, err)
	}
	return nil
}

// Listen accepts connections and calls handler for each message.
func Listen(ctx context.Context, port int, handler OnMessage) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
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
			log.Printf("[transport] accept: %v", err)
			continue
		}
		go handleConn(conn, handler)
	}
}

func handleConn(conn net.Conn, handler OnMessage) {
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	header := make([]byte, 5)
	if _, err := io.ReadFull(conn, header); err != nil {
		if err != io.EOF {
			log.Printf("[transport] read header: %v", err)
		}
		return
	}

	msgType := header[0]
	payloadLen := binary.BigEndian.Uint32(header[1:5])

	if msgType == MsgClipboard && payloadLen > MaxClipboardSize {
		log.Printf("[transport] clipboard too large: %d", payloadLen)
		return
	}
	if payloadLen > 10*1024*1024 {
		log.Printf("[transport] payload too large: %d", payloadLen)
		return
	}

	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(conn, payload); err != nil {
		log.Printf("[transport] read payload: %v", err)
		return
	}

	handler(msgType, payload)
}
