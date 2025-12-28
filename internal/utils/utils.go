// Package utils provides shared helper functions.
package utils

import (
	"os"
	"path/filepath"
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

// ExpandPath replaces "~" with the user's home directory and expands environment variables.
func ExpandPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	// Expand environment variables first
	expanded := os.ExpandEnv(path)

	// Handle tilde expansion
	if strings.HasPrefix(expanded, "~/") || expanded == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if expanded == "~" {
			return home, nil
		}
		return filepath.Join(home, expanded[2:]), nil
	}

	return expanded, nil
}
