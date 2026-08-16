//go:build windows

package clipboard

import (
	"log"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procOpenClipboard    = user32.NewProc("OpenClipboard")
	procCloseClipboard   = user32.NewProc("CloseClipboard")
	procGetClipboardData = user32.NewProc("GetClipboardData")
	procGlobalLock       = kernel32.NewProc("GlobalLock")
	procGlobalUnlock     = kernel32.NewProc("GlobalUnlock")
)

const (
	CF_HDROP = 15
)

// DROPFILES is the Win32 DROPFILES structure for CF_HDROP clipboard data.
// https://learn.microsoft.com/en-us/windows/win32/api/shlobj_core/ns-shlobj_core-dropfiles
type DROPFILES struct {
	pFiles uint32 // offset to file list from beginning of structure
	pt     struct {
		x, y int32
	}
	fNC   int32 // BOOL — non-client area flag
	fWide int32 // BOOL — if nonzero, file names are UTF-16
}

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

	ptr, _, _ := procGlobalLock.Call(h)
	if ptr == 0 {
		return nil
	}
	defer procGlobalUnlock.Call(h)

	df := (*DROPFILES)(unsafe.Pointer(ptr))

	// File list starts at offset pFiles from the DROPFILES structure.
	// unsafe.Add / unsafe.Slice are checkptr-aware, so this passes `go vet`
	// and won't trip the -checkptr instrumentation under `go test`.
	base := unsafe.Add(unsafe.Pointer(ptr), df.pFiles)

	var files []string
	if df.fWide != 0 {
		// UTF-16 encoded file names (terminated by a double null).
		off := uintptr(0)
		for {
			// Find null terminator.
			end := off
			for *(*uint16)(unsafe.Add(base, end)) != 0 {
				end += 2
			}
			nchars := int((end - off) / 2)
			if nchars == 0 {
				break // double null — end of list
			}
			// Decode the UTF-16 slice.
			raw := unsafe.Slice((*uint16)(unsafe.Add(base, off)), nchars)
			files = append(files, string(utf16.Decode(raw)))
			off = end + 2 // skip null terminator
		}
	} else {
		// ANSI file names (terminated by a double null).
		off := uintptr(0)
		for {
			end := off
			for *(*byte)(unsafe.Add(base, end)) != 0 {
				end++
			}
			nbytes := int(end - off)
			if nbytes == 0 {
				break // double null — end of list
			}
			raw := unsafe.Slice((*byte)(unsafe.Add(base, off)), nbytes)
			files = append(files, string(raw))
			off = end + 1 // skip null terminator
		}
	}

	if len(files) == 0 {
		return nil
	}
	return files
}
