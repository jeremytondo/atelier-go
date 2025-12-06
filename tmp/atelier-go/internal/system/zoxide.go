package system

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"
)

// GetZoxideEntries returns a list of frequent directories from zoxide.
func GetZoxideEntries() ([]string, error) {
	cmd := exec.Command("zoxide", "query", "-l")
	output, err := cmd.Output()
	if err != nil {
		// If zoxide is not installed or fails, return empty list
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
