// Package utils provides shared helper functions.
package utils

import (
	"fmt"
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

// GetCanonicalPath returns the absolute, case-corrected, and symlink-resolved version of a path.
// This is critical on macOS where the filesystem is case-insensitive but tools often expect
// the canonical disk casing. It performs a component-by-component lookup to ensure
// the returned path matches exactly what is on disk.
func GetCanonicalPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	// 1. Get absolute path
	abs, err := filepath.Abs(path)
	if err != nil {
		return path, err
	}

	// 2. Resolve symlinks
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return abs, nil // Return absolute path if resolution fails
	}

	// 3. Force correct casing (Recursive ReadDir lookup)
	return correctCasing(resolved)
}

func correctCasing(path string) (string, error) {
	if path == "/" || path == "" || path == "." {
		return path, nil
	}

	parent := filepath.Dir(path)
	if parent == path {
		return path, nil
	}

	canonicalParent, err := correctCasing(parent)
	if err != nil {
		return path, err
	}

	entries, err := os.ReadDir(canonicalParent)
	if err != nil {
		// If we can't read the directory, just return the path as-is
		return filepath.Join(canonicalParent, filepath.Base(path)), nil
	}

	base := filepath.Base(path)
	for _, entry := range entries {
		if strings.EqualFold(entry.Name(), base) {
			return filepath.Join(canonicalParent, entry.Name()), nil
		}
	}

	return filepath.Join(canonicalParent, base), nil
}

// SetTerminalTitle updates the terminal window title using ANSI escape sequences.
func SetTerminalTitle(title string) {
	fmt.Printf("\033]0;%s\007", title)
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
