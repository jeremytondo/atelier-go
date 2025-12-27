// Package utils provides shared helper functions.
package utils

import (
	"os"
	"strings"
)

// ShortenPath replaces the user's home directory with "~" in the given path.
// If the path is not within the home directory, it returns the original path.
func ShortenPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if after, found := strings.CutPrefix(path, home); found {
		return "~" + after
	}
	return path
}
