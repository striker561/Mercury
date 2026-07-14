package storage

import (
	"github.com/adrg/xdg"
)

// DefaultPath returns the platform-appropriate path for the database.
//   - Linux:   ~/.local/share/mercury/mercury.db
//   - macOS:   ~/Library/Application Support/mercury/mercury.db
//   - Windows: %APPDATA%/mercury/mercury.db
func DefaultPath() (string, error) {
	return xdg.DataFile("mercury/mercury.db")
}

// GetPassphrase returns the saved passphrase, or empty string.
func (d *DB) GetPassphrase() string {
	return d.GetDefaulted(KeyPassphrase)
}

// SetPassphrase saves the passphrase and marks sync as enabled.
func (d *DB) SetPassphrase(p string) {
	d.Set(KeyPassphrase, p)
	d.Set(KeySyncEnabled, "true")
}

// IsPaused returns true if sync was paused on last shutdown.
func (d *DB) IsPaused() bool {
	return d.GetDefaulted(KeyPaused) == "true"
}

// SetPaused saves the paused state.
func (d *DB) SetPaused(paused bool) {
	v := "false"
	if paused {
		v = "true"
	}
	d.Set(KeyPaused, v)
}

// IsFilesAllowed returns whether file transfers are accepted.
func (d *DB) IsFilesAllowed() bool {
	return d.GetDefaulted(KeyAllowFiles) == "true"
}

// SetFilesAllowed saves the file transfer preference.
func (d *DB) SetFilesAllowed(allowed bool) {
	v := "false"
	if allowed {
		v = "true"
	}
	d.Set(KeyAllowFiles, v)
}

// GetFolder returns the saved received-files folder path.
func (d *DB) GetFolder() string {
	return d.GetDefaulted(KeyReceivedFolder)
}

// SetFolder saves the received-files folder path.
func (d *DB) SetFolder(path string) {
	d.Set(KeyReceivedFolder, path)
}
