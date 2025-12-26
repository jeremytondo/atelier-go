package system

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExpandPath expands the tilde (~) in the beginning of a path to the user's home directory.
// If the path does not start with ~, it is returned as is.
// It also expands environment variables using os.ExpandEnv.
func ExpandPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	// Expand ~
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home dir: %w", err)
		}
		if path == "~" {
			path = home
		} else if strings.HasPrefix(path, "~/") {
			path = filepath.Join(home, path[2:])
		}
	}

	// Expand env vars
	return os.ExpandEnv(path), nil
}
