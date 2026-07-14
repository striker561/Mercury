package transfer

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
