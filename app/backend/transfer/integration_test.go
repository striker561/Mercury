package transfer

import (
	"bytes"
	"context"
	"crypto/rand"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"mercury/app/backend/crypto"
	"mercury/app/backend/transport"
)

const testPassphrase = "mercury-transfer-integration"

// testListener mirrors app.go: transport frames → decrypt → OnWireMessage.
type testListener struct {
	ctx    context.Context
	cancel context.CancelFunc
	addr   string
	mgr    *Manager
	key    []byte
}

func startTestListener(t *testing.T, recvMgr *Manager, key []byte) *testListener {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		cancel()
		t.Fatal(err)
	}

	go func() {
		_ = transport.Serve(ctx, ln, func(msgType byte, payload []byte) {
			dec, err := crypto.Decrypt(payload, key)
			if err != nil {
				return
			}
			recvMgr.OnWireMessage(msgType, dec)
		})
	}()

	return &testListener{
		ctx:    ctx,
		cancel: cancel,
		addr:   ln.Addr().String(),
		mgr:    recvMgr,
		key:    key,
	}
}

func (l *testListener) Close() {
	l.cancel()
}

func waitTransferDone(t *testing.T, m *Manager, tid string) *Progress {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if p := m.Progress(tid); p != nil {
			switch p.Status {
			case StatusDone, StatusFailed:
				return p
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("transfer %s did not finish within timeout", tid)
	return nil
}

func writeTempFile(t *testing.T, name string, data []byte) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func readReceivedFile(t *testing.T, dir, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestIntegrationSendFileForOffer(t *testing.T) {
	key := crypto.DeriveKey(testPassphrase)
	want := []byte("hello from integration test")
	src := writeTempFile(t, "payload.bin", want)

	recvMgr := NewManager(key)
	ln := startTestListener(t, recvMgr, key)
	defer ln.Close()

	offerID := "0123456789abcdef"
	recvMgr.IncomingOfferWithID(offerID, "payload.bin", int64(len(want)), "127.0.0.1")
	saveDir := t.TempDir()
	tid, err := recvMgr.AcceptOffer(offerID, saveDir)
	if err != nil {
		t.Fatal(err)
	}

	sendMgr := NewManager(key)
	sendTid, err := sendMgr.SendFileForOffer(offerID, ln.addr, src)
	if err != nil {
		t.Fatal(err)
	}

	if p := waitTransferDone(t, recvMgr, tid); p.Status != StatusDone {
		t.Fatalf("receive status %q, want done", p.Status)
	}
	if p := waitTransferDone(t, sendMgr, sendTid); p.Status != StatusDone {
		t.Fatalf("send status %q, want done", p.Status)
	}
	if p := recvMgr.Progress(tid); p.Received != int64(len(want)) {
		t.Fatalf("received %d bytes, want %d", p.Received, len(want))
	}
	if got := readReceivedFile(t, saveDir, "payload.bin"); !bytes.Equal(got, want) {
		t.Fatalf("file mismatch: got %q, want %q", got, want)
	}
}

func TestIntegrationMultiChunkFile(t *testing.T) {
	key := crypto.DeriveKey(testPassphrase)
	want := make([]byte, chunkSize+512)
	if _, err := rand.Read(want); err != nil {
		t.Fatal(err)
	}
	src := writeTempFile(t, "large.bin", want)

	recvMgr := NewManager(key)
	ln := startTestListener(t, recvMgr, key)
	defer ln.Close()

	offerID := "1111111111111111"
	recvMgr.IncomingOfferWithID(offerID, "large.bin", int64(len(want)), ln.addr)
	saveDir := t.TempDir()
	tid, err := recvMgr.AcceptOffer(offerID, saveDir)
	if err != nil {
		t.Fatal(err)
	}

	sendMgr := NewManager(key)
	sendTid, err := sendMgr.SendFileForOffer(offerID, ln.addr, src)
	if err != nil {
		t.Fatal(err)
	}

	if p := waitTransferDone(t, recvMgr, tid); p.Status != StatusDone {
		t.Fatalf("receive status %q, want done", p.Status)
	}
	if p := waitTransferDone(t, sendMgr, sendTid); p.Status != StatusDone {
		t.Fatalf("send status %q, want done", p.Status)
	}
	if got := readReceivedFile(t, saveDir, "large.bin"); !bytes.Equal(got, want) {
		t.Fatal("multi-chunk file content mismatch")
	}
}

func TestIntegrationConcurrentTransfers(t *testing.T) {
	key := crypto.DeriveKey(testPassphrase)
	wantA := bytes.Repeat([]byte("A"), chunkSize/4)
	wantB := bytes.Repeat([]byte("B"), chunkSize/4)
	srcA := writeTempFile(t, "a.bin", wantA)
	srcB := writeTempFile(t, "b.bin", wantB)

	recvMgr := NewManager(key)
	ln := startTestListener(t, recvMgr, key)
	defer ln.Close()

	offerA := "aaaaaaaaaaaaaaaa"
	offerB := "bbbbbbbbbbbbbbbb"
	recvMgr.IncomingOfferWithID(offerA, "a.bin", int64(len(wantA)), ln.addr)
	recvMgr.IncomingOfferWithID(offerB, "b.bin", int64(len(wantB)), ln.addr)
	saveDir := t.TempDir()
	tidA, err := recvMgr.AcceptOffer(offerA, saveDir)
	if err != nil {
		t.Fatal(err)
	}
	tidB, err := recvMgr.AcceptOffer(offerB, saveDir)
	if err != nil {
		t.Fatal(err)
	}

	sendMgr := NewManager(key)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		sendTid, err := sendMgr.SendFileForOffer(offerA, ln.addr, srcA)
		if err != nil {
			t.Error(err)
			return
		}
		if p := waitTransferDone(t, sendMgr, sendTid); p.Status != StatusDone {
			t.Errorf("send A status %q", p.Status)
		}
	}()
	go func() {
		defer wg.Done()
		sendTid, err := sendMgr.SendFileForOffer(offerB, ln.addr, srcB)
		if err != nil {
			t.Error(err)
			return
		}
		if p := waitTransferDone(t, sendMgr, sendTid); p.Status != StatusDone {
			t.Errorf("send B status %q", p.Status)
		}
	}()
	wg.Wait()

	if p := waitTransferDone(t, recvMgr, tidA); p.Status != StatusDone {
		t.Fatalf("receive A status %q", p.Status)
	}
	if p := waitTransferDone(t, recvMgr, tidB); p.Status != StatusDone {
		t.Fatalf("receive B status %q", p.Status)
	}
	if got := readReceivedFile(t, saveDir, "a.bin"); !bytes.Equal(got, wantA) {
		t.Fatal("file A content mismatch")
	}
	if got := readReceivedFile(t, saveDir, "b.bin"); !bytes.Equal(got, wantB) {
		t.Fatal("file B content mismatch")
	}
}

func TestIntegrationWrongKeyRejected(t *testing.T) {
	recvKey := crypto.DeriveKey("receiver-passphrase")
	sendKey := crypto.DeriveKey("sender-passphrase")

	recvMgr := NewManager(recvKey)
	ln := startTestListener(t, recvMgr, recvKey)
	defer ln.Close()

	offerID := "deadbeefdeadbeef"
	ch := make(chan []byte, 2)
	recvMgr.receiveCh[offerID] = ch

	conn, err := net.Dial("tcp", ln.addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	wire, err := packChunkPayload(offerID, []byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	enc, err := crypto.Encrypt(wire, sendKey)
	if err != nil {
		t.Fatal(err)
	}
	if err := transport.WriteFrame(conn, transport.MsgFileChunk, enc); err != nil {
		t.Fatal(err)
	}

	time.Sleep(50 * time.Millisecond)
	select {
	case <-ch:
		t.Fatal("chunk encrypted with wrong key should not be delivered")
	default:
	}
}

func TestIntegrationTamperedCiphertextDropped(t *testing.T) {
	key := crypto.DeriveKey(testPassphrase)
	recvMgr := NewManager(key)
	ln := startTestListener(t, recvMgr, key)
	defer ln.Close()

	offerID := "abababababababab"
	ch := make(chan []byte, 2)
	recvMgr.receiveCh[offerID] = ch

	wire, _ := packChunkPayload(offerID, []byte("data"))
	enc, err := crypto.Encrypt(wire, key)
	if err != nil {
		t.Fatal(err)
	}
	enc[len(enc)-1] ^= 0xff // corrupt auth tag

	conn, err := net.Dial("tcp", ln.addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := transport.WriteFrame(conn, transport.MsgFileChunk, enc); err != nil {
		t.Fatal(err)
	}

	time.Sleep(50 * time.Millisecond)
	select {
	case <-ch:
		t.Fatal("tampered ciphertext should not be delivered")
	default:
	}
}

func TestIntegrationEarlyEndMarkerFails(t *testing.T) {
	key := crypto.DeriveKey(testPassphrase)
	recvMgr := NewManager(key)
	ln := startTestListener(t, recvMgr, key)
	defer ln.Close()

	offerID := "cafebabecafebabe"
	wantSize := int64(1024)
	recvMgr.IncomingOfferWithID(offerID, "early.bin", wantSize, ln.addr)
	saveDir := t.TempDir()
	tid, err := recvMgr.AcceptOffer(offerID, saveDir)
	if err != nil {
		t.Fatal(err)
	}

	conn, err := net.Dial("tcp", ln.addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Send one small chunk then end marker before file is complete.
	chunk, err := packChunkPayload(offerID, []byte("partial"))
	if err != nil {
		t.Fatal(err)
	}
	enc, err := crypto.Encrypt(chunk, key)
	if err != nil {
		t.Fatal(err)
	}
	if err := transport.WriteFrame(conn, transport.MsgFileChunk, enc); err != nil {
		t.Fatal(err)
	}
	endEnc, err := crypto.Encrypt([]byte(offerID), key)
	if err != nil {
		t.Fatal(err)
	}
	if err := transport.WriteFrame(conn, transport.MsgFileEnd, endEnc); err != nil {
		t.Fatal(err)
	}

	if p := waitTransferDone(t, recvMgr, tid); p.Status != StatusFailed {
		t.Fatalf("status %q, want failed", p.Status)
	}
	if _, err := os.Stat(filepath.Join(saveDir, "early.bin")); !os.IsNotExist(err) {
		t.Fatal("partial file should be removed after early end")
	}
}

func TestIntegrationEmptyFile(t *testing.T) {
	key := crypto.DeriveKey(testPassphrase)
	src := writeTempFile(t, "empty.bin", nil)

	recvMgr := NewManager(key)
	ln := startTestListener(t, recvMgr, key)
	defer ln.Close()

	offerID := "0000000000000000"
	recvMgr.IncomingOfferWithID(offerID, "empty.bin", 0, ln.addr)
	saveDir := t.TempDir()
	tid, err := recvMgr.AcceptOffer(offerID, saveDir)
	if err != nil {
		t.Fatal(err)
	}

	sendMgr := NewManager(key)
	sendTid, err := sendMgr.SendFileForOffer(offerID, ln.addr, src)
	if err != nil {
		t.Fatal(err)
	}

	if p := waitTransferDone(t, recvMgr, tid); p.Status != StatusDone {
		t.Fatalf("receive status %q", p.Status)
	}
	if p := waitTransferDone(t, sendMgr, sendTid); p.Status != StatusDone {
		t.Fatalf("send status %q", p.Status)
	}
	info, err := os.Stat(filepath.Join(saveDir, "empty.bin"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 0 {
		t.Fatalf("empty file size %d", info.Size())
	}
}
