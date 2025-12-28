// Package utils provides shared helper functions.
package utils

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
)

// NewTableWriter creates a configured tabwriter for consistent table output.
func NewTableWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
}

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

// GetHostname returns the effective hostname for configuration purposes.
// It prioritizes the ATELIER_HOSTNAME environment variable.
// If not set, it falls back to the system hostname, taking only the part before the first dot.
// The result is always lowercased.
func GetHostname() (string, error) {
	if h := os.Getenv("ATELIER_HOSTNAME"); h != "" {
		return strings.ToLower(h), nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	// Split by dot and take the first part
	if i := strings.Index(hostname, "."); i != -1 {
		hostname = hostname[:i]
	}

	return strings.ToLower(hostname), nil
}
