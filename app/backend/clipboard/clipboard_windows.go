//go:build windows

package clipboard

import (
	"log"
	"syscall"
	"unsafe"
)

var (
	user32  = syscall.NewLazyDLL("user32.dll")
	shell32 = syscall.NewLazyDLL("shell32.dll")

	procOpenClipboard    = user32.NewProc("OpenClipboard")
	procCloseClipboard   = user32.NewProc("CloseClipboard")
	procGetClipboardData = user32.NewProc("GetClipboardData")
	procDragQueryFileW   = shell32.NewProc("DragQueryFileW")
)

const CF_HDROP = 15

// readFileURLs reads file paths from the Windows clipboard (CF_HDROP format).
// Returns nil if no file paths are present.
func readFileURLs() []string {
	// The clipboard can hold arbitrary or malformed data — never let parsing
	// quirks crash the watcher goroutine (a Go panic in any goroutine would
	// terminate the whole process).
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[clipboard] readFileURLs recovered: %v", r)
		}
	}()

	r, _, _ := procOpenClipboard.Call(0)
	if r == 0 {
		return nil
	}
	defer procCloseClipboard.Call()

	h, _, _ := procGetClipboardData.Call(CF_HDROP)
	if h == 0 {
		return nil
	}

	count, _, _ := procDragQueryFileW.Call(h, 0xFFFFFFFF, 0, 0)
	if count == 0 || count > 10000 {
		return nil
	}

	files := make([]string, 0, int(count))
	for i := uintptr(0); i < count; i++ {
		n, _, _ := procDragQueryFileW.Call(h, i, 0, 0)
		if n == 0 {
			continue
		}
		buf := make([]uint16, int(n)+1)
		procDragQueryFileW.Call(h, i, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
		files = append(files, syscall.UTF16ToString(buf))
	}
	if len(files) == 0 {
		return nil
	}
	return files
}
