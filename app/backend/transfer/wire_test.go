package transfer

import "testing"

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
	chA := make(chan []byte, 2)
	chB := make(chan []byte, 2)
	m.receiveCh[offerA] = chA
	m.receiveCh[offerB] = chB

	wireA, _ := packChunkPayload(offerA, []byte("aaa"))
	m.OnWireMessage(1, wireA)

	select {
	case got := <-chA:
		if string(got) != "aaa" {
			t.Fatalf("chA got %q", got)
		}
	default:
		t.Fatal("expected chunk on chA")
	}

	select {
	case <-chB:
		t.Fatal("unexpected chunk on chB")
	default:
	}
}

func TestOnWireMessageEndMarker(t *testing.T) {
	m := NewManager(nil)
	offerID := "cccccccccccccccc"
	ch := make(chan []byte, 1)
	m.receiveCh[offerID] = ch

	m.OnWireMessage(2, []byte(offerID))

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("channel should be closed, not deliver data")
		}
	default:
		t.Fatal("expected channel to be closed")
	}
}
