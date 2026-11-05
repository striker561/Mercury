package system

import (
	"errors"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/updater"
)

func TestUpdateResultLabel(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		state updater.State
		want  string
	}{
		{name: "up to date", state: updater.StateUpToDate, want: "Up to date"},
		{name: "ready", state: updater.StateReady, want: "Update ready — restart Mercury"},
		{name: "error first line", err: errors.New("boom\nresponse body"), want: "Update failed: boom"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := updateResultLabel(tt.err, tt.state); got != tt.want {
				t.Fatalf("updateResultLabel() = %q, want %q", got, tt.want)
			}
		})
	}
}
