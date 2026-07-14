//go:build !darwin

package clipboard

// readFileURLs is a no-op on non-macOS platforms.  On macOS this uses
// CGO to read NSFilenamesPboardType from the general pasteboard.
func readFileURLs() []string {
	return nil
}
