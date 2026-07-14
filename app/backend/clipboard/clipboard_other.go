//go:build !darwin && !windows

package clipboard

// readFileURLs is a no-op on Linux and other non-macOS/non-Windows platforms.
// On macOS this uses CGO to read NSFilenamesPboardType from the general
// pasteboard. On Windows this reads CF_HDROP via Win32 API.
func readFileURLs() []string {
	return nil
}
