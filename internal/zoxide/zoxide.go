package zoxide

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Query returns a list of frequently used directories from zoxide.
// It executes `zoxide query -l` and parses the output.
func Query() ([]string, error) {
	cmd := exec.Command("zoxide", "query", "-l")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run zoxide query: %w", err)
	}

	var paths []string
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		path := strings.TrimSpace(scanner.Text())
		if path != "" {
			paths = append(paths, path)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse zoxide output: %w", err)
	}

	return paths, nil
}
