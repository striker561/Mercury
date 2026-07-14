// Package fileinfo detects whether clipboard text refers to a local file,
// resolves the real path (handling file:// URIs from macOS/Linux), and
// identifies basic file type categories.
package fileinfo

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Category represents a broad file type.
type Category int

const (
	Unknown Category = iota
	Image
	Video
	Audio
	Document
	Archive
	Code
	Executable
)

func (c Category) String() string {
	switch c {
	case Image:
		return "image"
	case Video:
		return "video"
	case Audio:
		return "audio"
	case Document:
		return "document"
	case Archive:
		return "archive"
	case Code:
		return "code"
	case Executable:
		return "executable"
	default:
		return "unknown"
	}
}

// Info describes a file detected on the local filesystem.
type Info struct {
	Path     string   // absolute path on disk
	Name     string   // base filename
	Size     int64    // file size in bytes
	Ext      string   // lowercase extension including dot, e.g. ".png"
	Category Category // broad type category
}

// Detect checks whether text looks like a local file path and returns
// file info if the file exists.  Returns nil if text doesn't point to
// a valid file.
func Detect(text string) *Info {
	p := resolve(text)
	if p == "" {
		return nil
	}
	fi, err := os.Stat(p)
	if err != nil || fi.IsDir() {
		return nil
	}
	return &Info{
		Path:     p,
		Name:     filepath.Base(p),
		Size:     fi.Size(),
		Ext:      strings.ToLower(filepath.Ext(p)),
		Category: classify(filepath.Ext(p)),
	}
}

// resolve converts clipboard text to an absolute file path, handling
// file:// URIs from macOS Finder and Linux file managers.
func resolve(text string) string {
	s := strings.TrimSpace(text)
	if s == "" {
		return ""
	}
	// macOS Finder and Linux Nautilus often copy file:// URIs.
	if strings.HasPrefix(s, "file://") {
		u, err := url.Parse(s)
		if err != nil {
			return ""
		}
		return u.Path
	}
	return s
}

// classify maps a file extension to a broad category.
func classify(ext string) Category {
	switch strings.ToLower(ext) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".tiff", ".tif",
		".svg", ".ico", ".heic", ".avif":
		return Image
	case ".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv", ".webm", ".m4v":
		return Video
	case ".mp3", ".wav", ".flac", ".aac", ".ogg", ".wma", ".m4a":
		return Audio
	case ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
		".txt", ".rtf", ".md", ".csv", ".odt", ".ods":
		return Document
	case ".zip", ".tar", ".gz", ".bz2", ".xz", ".7z", ".rar", ".tgz":
		return Archive
	case ".go", ".js", ".ts", ".py", ".java", ".rs", ".c", ".cpp", ".h",
		".css", ".html", ".json", ".xml", ".yaml", ".yml", ".toml",
		".sh", ".bash", ".zsh", ".rb", ".php", ".swift", ".kt", ".scala":
		return Code
	case ".exe", ".bin", ".app", ".deb", ".rpm", ".dmg", ".pkg":
		return Executable
	default:
		return Unknown
	}
}

// IsImage is a convenience helper.
func IsImage(ext string) bool {
	return classify(ext) == Image
}
