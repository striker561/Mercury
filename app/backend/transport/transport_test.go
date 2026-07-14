package transport

import (
	"bytes"
	"net"
	"testing"
)

func TestWriteFrameMultipleOnConnection(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	done := make(chan struct{})
	var gotTypes []byte
	var gotPayloads [][]byte

	go func() {
		defer close(done)
		conn, err := ln.Accept()
		if err != nil {
			t.Errorf("accept: %v", err)
			return
		}
		handleConn(conn, func(msgType byte, payload []byte) {
			gotTypes = append(gotTypes, msgType)
			gotPayloads = append(gotPayloads, bytes.Clone(payload))
		})
	}()

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	if err := WriteFrame(conn, MsgFileChunk, []byte("chunk-a")); err != nil {
		t.Fatal(err)
	}
	if err := WriteFrame(conn, MsgFileChunk, []byte("chunk-b")); err != nil {
		t.Fatal(err)
	}
	if err := WriteFrame(conn, MsgFileEnd, []byte("end")); err != nil {
		t.Fatal(err)
	}
	conn.Close()

	<-done

	if len(gotTypes) != 3 {
		t.Fatalf("got %d frames, want 3", len(gotTypes))
	}
	wantTypes := []byte{MsgFileChunk, MsgFileChunk, MsgFileEnd}
	if !bytes.Equal(gotTypes, wantTypes) {
		t.Fatalf("types %v, want %v", gotTypes, wantTypes)
	}
	if string(gotPayloads[0]) != "chunk-a" || string(gotPayloads[1]) != "chunk-b" || string(gotPayloads[2]) != "end" {
		t.Fatalf("payloads mismatch: %q, %q, %q", gotPayloads[0], gotPayloads[1], gotPayloads[2])
	}
}
