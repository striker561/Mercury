package transfer

import "testing"

func TestPeerTCPAddr(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"192.168.1.10", "192.168.1.10:47821"},
		{"192.168.1.10:47821", "192.168.1.10:47821"},
		{"127.0.0.1:54321", "127.0.0.1:54321"},
	}
	for _, tc := range tests {
		if got := peerTCPAddr(tc.in); got != tc.want {
			t.Errorf("peerTCPAddr(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
