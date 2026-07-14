//go:build darwin

package clipboard

/*
#cgo CFLAGS: -x objective-c -mmacosx-version-min=10.13
#cgo LDFLAGS: -framework Foundation -framework Cocoa

#import <Cocoa/Cocoa.h>
#import <stdlib.h>

// readFileURLsFromPasteboard reads POSIX file paths from the general
// pasteboard by querying for NSString objects (which NSPasteboard
// returns for NSFilenamesPboardType / NSPasteboardTypeFileURL).
//
// Returns a newline-separated C string, or NULL if no file URLs
// are on the pasteboard. Caller must free with free().
const char* readFileURLsFromPasteboard() {
    @autoreleasepool {
        NSPasteboard *pb = [NSPasteboard generalPasteboard];

        // Try modern readObjectsForClasses: first (macOS 10.6+)
        NSArray<NSString *> *accepted = @[NSString.class];
        NSDictionary *options = @{};
        NSArray *paths = [pb readObjectsForClasses:accepted options:options];

        // Filter: only return actual file paths (not random strings)
        NSMutableArray *files = [NSMutableArray array];
        for (NSString *s in paths) {
            if ([s hasPrefix:@"/"]) {
                [files addObject:s];
            }
        }

        if (files.count == 0) {
            // Fallback: try legacy NSFilenamesPboardType
            id plist = [pb propertyListForType:@"NSFilenamesPboardType"];
            if ([plist isKindOfClass:[NSArray class]]) {
                files = plist;
            }
        }

        if (files.count == 0) {
            return NULL;
        }

        NSString *joined = [files componentsJoinedByString:@"\n"];
        return strdup([joined UTF8String]);
    }
}
*/
import "C"
import (
	"strings"
	"unsafe"
)

// readFileURLs returns file paths from the macOS pasteboard.
// Returns nil if no file paths are present.
func readFileURLs() []string {
	cstr := C.readFileURLsFromPasteboard()
	if cstr == nil {
		return nil
	}
	defer C.free(unsafe.Pointer(cstr))

	raw := C.GoString(cstr)
	parts := strings.Split(raw, "\n")
	var out []string
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
