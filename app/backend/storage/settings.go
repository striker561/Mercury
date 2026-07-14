package storage

// All setting keys and their defaults live here.
// Adding a new setting? Add a key constant and an entry in defaults.
// No migration needed — missing keys return the default value automatically.

const (
	KeyPassphrase     = "passphrase"
	KeySyncEnabled    = "sync_enabled"
	KeyPaused         = "paused"
	KeyAllowFiles     = "allow_files"
	KeyReceivedFolder = "received_folder"
	KeyAutostart      = "autostart"
	KeyAutoAccept     = "auto_accept"
)

// defaultValues holds the fallback for every known setting.
var defaultValues = map[string]string{
	KeyPassphrase:     "",
	KeySyncEnabled:    "false",
	KeyPaused:         "false",
	KeyAllowFiles:     "true",
	KeyReceivedFolder: "~/Mercury/",
	KeyAutostart:      "false",
	KeyAutoAccept:     "false",
}

// AllKeys returns every registered setting key.
func AllKeys() []string {
	keys := make([]string, 0, len(defaultValues))
	for k := range defaultValues {
		keys = append(keys, k)
	}
	return keys
}

// Default returns the default value for a setting key.
// Returns empty string if the key isn't registered.
func Default(key string) string {
	return defaultValues[key]
}

// GetDefaulted retrieves a setting, falling back to its default if empty.
func (d *DB) GetDefaulted(key string) string {
	v, _ := d.Get(key)
	if v == "" {
		return Default(key)
	}
	return v
}

// All reads every row from the settings table in a single query and returns
// a map merged with defaults — any key not in the DB gets its default value.
//
// Why one query instead of N individual Get(key) calls?
// Each Get(key) is a separate SQL round-trip. With 30 settings that's 30
// round-trips. One SELECT * is always faster than 30 SELECT WHERE, even with
// SQLite. This matters less on a local DB than on a network one, but the
// principle is the same: batch reads, minimise round-trips.
func (d *DB) All() map[string]string {
	out := make(map[string]string, len(defaultValues))

	rows, err := d.conn.Query("SELECT key, value FROM settings")
	if err != nil {
		// If the query fails, fall back to pure defaults.
		goto defaults
	}
	defer rows.Close()

	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			continue
		}
		out[k] = v
	}

defaults:
	// Fill in defaults for any key not in the DB.
	for k, v := range defaultValues {
		if _, ok := out[k]; !ok {
			out[k] = v
		}
	}
	return out
}
