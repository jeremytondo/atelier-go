package system

import (
	"os"
	"strings"
)

// IsLocal checks if the provided host resolves to the local machine.
func IsLocal(host string) bool {
	if host == "localhost" || host == "127.0.0.1" || host == "0.0.0.0" || host == "" {
		return true
	}
	hostname, err := os.Hostname()
	if err == nil && strings.EqualFold(host, hostname) {
		return true
	}
	return false
}
