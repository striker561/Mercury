// Package storage provides a simple key-value store backed by SQLite.
// Used to persist Mercury settings between restarts.
package storage

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// DB wraps the SQLite connection.
type DB struct {
	conn *sql.DB
}

// Open opens (or creates) the SQLite database at path and runs migrations.
func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("storage open: %w", err)
	}

	// Enable WAL mode for better concurrency.
	if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("storage wal: %w", err)
	}

	// Create the settings table if it doesn't exist.
	if _, err := conn.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)
	`); err != nil {
		return nil, fmt.Errorf("storage migrate: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close closes the database.
func (d *DB) Close() error {
	return d.conn.Close()
}

// Get retrieves a setting by key. Returns empty string if not found.
func (d *DB) Get(key string) (string, error) {
	var value string
	err := d.conn.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("storage get %s: %w", key, err)
	}
	return value, nil
}

// Set saves a setting by key. Inserts or updates if the key exists.
func (d *DB) Set(key, value string) error {
	_, err := d.conn.Exec(
		"INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
		key, value,
	)
	if err != nil {
		return fmt.Errorf("storage set %s: %w", key, err)
	}
	return nil
}
