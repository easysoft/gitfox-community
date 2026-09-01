// Package netutil provides network utility functions.
package netutil

import "net"

// LocalIP returns the first non-loopback IPv4 address of the host machine.
// If no non-loopback address is found, it falls back to "127.0.0.1".
func LocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}

	return "127.0.0.1"
}
