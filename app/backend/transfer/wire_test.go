package transfer

import (
	"testing"
	"time"
)

func newTestReceiveStream() *receiveStream {
	return &receiveStream{
		chunks: make(chan []byte, 2),
		end:    make(chan struct{}),
		done:   make(chan struct{}),
	}
}

func TestPackChunkPayload(t *testing.T) {
	offerID := "0123456789abcdef"
	data := []byte("hello")

	wire, err := packChunkPayload(offerID, data)
	if err != nil {
		t.Fatal(err)
	}
	if len(wire) != offerIDLen+len(data) {
		t.Fatalf("wire len %d, want %d", len(wire), offerIDLen+len(data))
	}
	gotID, err := offerIDFromPayload(wire)
	if err != nil {
		t.Fatal(err)
	}
	if gotID != offerID {
		t.Fatalf("offer ID %q, want %q", gotID, offerID)
	}
	if string(wire[offerIDLen:]) != "hello" {
		t.Fatalf("chunk data %q", wire[offerIDLen:])
	}
}

func TestOnWireMessageRoutesByOfferID(t *testing.T) {
	m := NewManager(nil)
	offerA := "aaaaaaaaaaaaaaaa"
	offerB := "bbbbbbbbbbbbbbbb"
	streamA := newTestReceiveStream()
	streamB := newTestReceiveStream()
	m.receiveCh[offerA] = streamA
	m.receiveCh[offerB] = streamB

	wireA, _ := packChunkPayload(offerA, []byte("aaa"))
	m.OnWireMessage(1, wireA)

	select {
	case got := <-streamA.chunks:
		if string(got) != "aaa" {
			t.Fatalf("streamA got %q", got)
		}
	default:
		t.Fatal("expected chunk on streamA")
	}

	select {
	case <-streamB.chunks:
		t.Fatal("unexpected chunk on streamB")
	default:
	}
}

func TestOnWireMessageEndMarker(t *testing.T) {
	m := NewManager(nil)
	offerID := "cccccccccccccccc"
	stream := newTestReceiveStream()
	m.receiveCh[offerID] = stream

	m.OnWireMessage(2, []byte(offerID))

	select {
	case <-stream.end:
	default:
		t.Fatal("expected the end marker to close the stream")
	}
}

func TestOnWireMessageBlocksInsteadOfDropping(t *testing.T) {
	m := NewManager(nil)
	offerID := "dddddddddddddddd"
	stream := newTestReceiveStream()
	stream.chunks = make(chan []byte, 1)
	m.receiveCh[offerID] = stream

	wire, _ := packChunkPayload(offerID, []byte("first"))
	m.OnWireMessage(1, wire)

	done := make(chan struct{})
	go func() {
		m.OnWireMessage(1, wire)
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("a full stream must apply backpressure, not drop")
	case <-time.After(50 * time.Millisecond):
	}

	<-stream.chunks

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("blocked send did not resume after the chunk was consumed")
	}
}
