package system

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"
)

// GetSessions returns a list of active shpool sessions.
func GetSessions() ([]string, error) {
	cmd := exec.Command("shpool", "list")
	output, err := cmd.Output()
	if err != nil {
		// If shpool is not installed or fails, we return an empty list gracefully
		// You might want to log this in a real app, but for now we just return empty
		return []string{}, nil
	}

	var sessions []string
	scanner := bufio.NewScanner(bytes.NewReader(output))

	// Skip header
	if scanner.Scan() {
		// First line is header, ignore
	}

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) > 0 {
			sessions = append(sessions, fields[0])
		}
	}

	return sessions, nil
}
