package app

import "testing"

func TestNeedsCairoRenderer(t *testing.T) {
	tests := []struct {
		name           string
		goos           string
		gskRenderer    string
		sessionType    string
		waylandDisplay string
		hasNVIDIA      bool
		force          bool
		want           bool
	}{
		{name: "cosmic nvidia wayland", goos: "linux", sessionType: "wayland", hasNVIDIA: true, want: true},
		{name: "pop nvidia wayland", goos: "linux", waylandDisplay: "wayland-0", hasNVIDIA: true, want: true},
		{name: "explicit renderer wins", goos: "linux", gskRenderer: "gl", sessionType: "wayland", hasNVIDIA: true, want: false},
		{name: "gnome nvidia wayland", goos: "linux", sessionType: "wayland", hasNVIDIA: true, want: true},
		{name: "kde nvidia wayland", goos: "linux", sessionType: "wayland", hasNVIDIA: true, want: true},
		{name: "x11 session", goos: "linux", sessionType: "x11", hasNVIDIA: true, want: false},
		{name: "no nvidia", goos: "linux", sessionType: "wayland", want: false},
		{name: "not linux", goos: "darwin", sessionType: "wayland", hasNVIDIA: true, want: false},
		{name: "force overrides explicit renderer", goos: "linux", gskRenderer: "gl", sessionType: "wayland", hasNVIDIA: true, force: true, want: true},
		{name: "force on x11", goos: "linux", sessionType: "x11", force: true, want: true},
		{name: "force ignored off linux", goos: "darwin", force: true, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := needsCairoRenderer(tt.goos, tt.gskRenderer, tt.sessionType, tt.waylandDisplay, tt.hasNVIDIA, tt.force); got != tt.want {
				t.Fatalf("needsCairoRenderer() = %v, want %v", got, tt.want)
			}
		})
	}
}
