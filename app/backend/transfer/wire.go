package transfer

import "fmt"

const offerIDLen = 16 // hex-encoded 8 random bytes from NewOfferID()

func packChunkPayload(offerID string, data []byte) ([]byte, error) {
	if len(offerID) != offerIDLen {
		return nil, fmt.Errorf("transfer: offer ID length %d, want %d", len(offerID), offerIDLen)
	}
	out := make([]byte, offerIDLen+len(data))
	copy(out[:offerIDLen], offerID)
	copy(out[offerIDLen:], data)
	return out, nil
}

func offerIDFromPayload(payload []byte) (string, error) {
	if len(payload) < offerIDLen {
		return "", fmt.Errorf("transfer: payload too short (%d)", len(payload))
	}
	return string(payload[:offerIDLen]), nil
}
