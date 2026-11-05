package app

import "testing"

func TestNeedsX11Backend(t *testing.T) {
	tests := []struct {
		name           string
		goos           string
		gdkBackend     string
		sessionType    string
		waylandDisplay string
		desktop        string
		hasNVIDIA      bool
		want           bool
	}{
		{name: "cosmic nvidia wayland", goos: "linux", sessionType: "wayland", desktop: "COSMIC", hasNVIDIA: true, want: true},
		{name: "cosmic nvidia wayland display only", goos: "linux", waylandDisplay: "wayland-0", desktop: "cosmic", hasNVIDIA: true, want: true},
		{name: "explicit backend wins", goos: "linux", gdkBackend: "wayland", sessionType: "wayland", desktop: "COSMIC", hasNVIDIA: true, want: false},
		{name: "x11 session", goos: "linux", sessionType: "x11", desktop: "COSMIC", hasNVIDIA: true, want: false},
		{name: "no nvidia", goos: "linux", sessionType: "wayland", desktop: "COSMIC", want: false},
		{name: "gnome untouched", goos: "linux", sessionType: "wayland", desktop: "GNOME", hasNVIDIA: true, want: false},
		{name: "kde untouched", goos: "linux", sessionType: "wayland", desktop: "KDE", hasNVIDIA: true, want: false},
		{name: "not linux", goos: "darwin", sessionType: "wayland", desktop: "COSMIC", hasNVIDIA: true, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := needsX11Backend(tt.goos, tt.gdkBackend, tt.sessionType, tt.waylandDisplay, tt.desktop, tt.hasNVIDIA); got != tt.want {
				t.Fatalf("needsX11Backend() = %v, want %v", got, tt.want)
			}
		})
	}
}
