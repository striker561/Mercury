package storage

import (
	"github.com/adrg/xdg"
)

// DefaultPath returns the platform-appropriate path for the database.
//   - Linux:   ~/.local/share/mercury/mercury.db
//   - macOS:   ~/Library/Application Support/mercury/mercury.db
//   - Windows: %APPDATA%/mercury/mercury.db
//
// Named mercury.db so the same file can host future tables (transfer history, etc.).
func DefaultPath() (string, error) {
	return xdg.DataFile("mercury/mercury.db")
}

// GetPassphrase returns the saved passphrase, or empty string.
func (d *DB) GetPassphrase() string {
	v, _ := d.Get("passphrase")
	return v
}

// SetPassphrase saves the passphrase and marks sync as enabled.
func (d *DB) SetPassphrase(p string) {
	d.Set("passphrase", p)
	d.Set("sync_enabled", "true")
}

// IsPaused returns true if sync was paused on last shutdown.
func (d *DB) IsPaused() bool {
	v, _ := d.Get("paused")
	return v == "true"
}

// SetPaused saves the paused state.
func (d *DB) SetPaused(paused bool) {
	v := "false"
	if paused {
		v = "true"
	}
	d.Set("paused", v)
}

// GetFolder returns the saved received-files folder path.
func (d *DB) GetFolder() string {
	v, _ := d.Get("received_folder")
	if v == "" {
		return "~/Mercury/"
	}
	return v
}

// SetFolder saves the received-files folder path.
func (d *DB) SetFolder(path string) {
	d.Set("received_folder", path)
}
