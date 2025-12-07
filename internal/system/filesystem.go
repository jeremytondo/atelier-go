package system

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"strings"
)

// GetAllEntries returns a list of all directories from the home directory using fd.
// It excludes .git and node_modules.
func GetAllEntries() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return []string{}, err
	}

	// fd --type d --base-directory $HOME --hidden --exclude .git --exclude node_modules
	cmd := exec.Command("fd", "--type", "d", "--base-directory", home, "--hidden", "--exclude", ".git", "--exclude", "node_modules")
	output, err := cmd.Output()
	if err != nil {
		// If fd fails or is not found, we return an empty list.
		// In a production app we might want to distinguish between "not found" and "error",
		// but per instructions we assume fd is required/available or we fail gracefully.
		return []string{}, nil
	}

	var paths []string
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			paths = append(paths, line)
		}
	}

	return paths, nil
}
