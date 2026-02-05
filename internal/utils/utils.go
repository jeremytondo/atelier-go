// Package utils provides shared helper functions.
package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/tabwriter"
)

var sanitizeRegex = regexp.MustCompile(`[^a-z0-9]+`)

// Sanitize cleans a string to be used as a session name component.
func Sanitize(s string) string {
	s = strings.ToLower(s)
	s = sanitizeRegex.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// NewTableWriter creates a configured tabwriter for consistent table output.
func NewTableWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
}

// RenderTable prints a tab-separated table to the provided writer.
func RenderTable(w io.Writer, headers []string, rows [][]string) error {
	tw := NewTableWriter(w)
	if _, err := fmt.Fprintf(tw, "%s\n", strings.Join(headers, "\t")); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	for _, row := range rows {
		if _, err := fmt.Fprintf(tw, "%s\n", strings.Join(row, "\t")); err != nil {
			return fmt.Errorf("error writing row: %w", err)
		}
	}

	if err := tw.Flush(); err != nil {
		return fmt.Errorf("error flushing writer: %w", err)
	}
	return nil
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

// GetCanonicalPath returns the absolute and symlink-resolved version of a path.
func GetCanonicalPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	// Get absolute path
	abs, err := filepath.Abs(path)
	if err != nil {
		return path, err
	}

	// Resolve symlinks
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return abs, nil // Return absolute path if resolution fails
	}

	return resolved, nil
}

// IconSSH is the Nerd Font icon used for SSH sessions (nf-fa-server).
const IconSSH = "\uf233"

// IsSSH returns true if the current process is running over an SSH connection.
func IsSSH() bool {
	return os.Getenv("SSH_CONNECTION") != "" || os.Getenv("SSH_CLIENT") != "" || os.Getenv("SSH_TTY") != ""
}

// SetTerminalTitle updates the terminal window title using ANSI escape sequences.
func SetTerminalTitle(title string) {
	if IsSSH() {
		title = IconSSH + " " + title
	}
	fmt.Printf("\033]0;%s\007", title)
}

// GetStateDir returns the XDG state directory for atelier-go.
// Creates the directory if it doesn't exist.
func GetStateDir() (string, error) {
	stateHome := os.Getenv("XDG_STATE_HOME")
	if stateHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		stateHome = filepath.Join(home, ".local", "state")
	}
	dir := filepath.Join(stateHome, "atelier-go", "sessions")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}
