package transfer

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"runtime"

	"mercury/app/backend/transport"
)

// stripPort returns the host portion of "host:port", or addr unchanged if
// no port is found.
func stripPort(addr string) string {
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
	}
	return addr
}

// peerTCPAddr returns the dial address for file transfer. If peerAddr already
// includes a port (e.g. "192.168.1.5:47821"), that port is used; otherwise
// transport.Port is appended to the host.
func peerTCPAddr(peerAddr string) string {
	if host, port, err := net.SplitHostPort(peerAddr); err == nil && port != "" {
		return net.JoinHostPort(host, port)
	}
	return fmt.Sprintf("%s:%d", stripPort(peerAddr), transport.Port)
}

// fileError adds the System Settings path a macOS user needs after a TCC denial.
func fileError(op string, err error) string {
	if runtime.GOOS == "darwin" && errors.Is(err, fs.ErrPermission) {
		return fmt.Sprintf("%s: %v — allow Mercury in System Settings → Privacy & Security → Files and Folders", op, err)
	}
	return fmt.Sprintf("%s: %v", op, err)
}
