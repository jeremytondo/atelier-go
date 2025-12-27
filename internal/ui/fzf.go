package ui

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Select opens fzf with the provided items and returns the selected item and the key pressed.
// If expects is provided, fzf will print the key pressed as the first line.
func Select(items []string, header string, prompt string, expects []string) (string, string, error) {
	args := []string{"--ansi", "--no-sort", "--layout=reverse", "--bind=esc:abort"}
	if header != "" {
		args = append(args, fmt.Sprintf("--header=%s", header))
	}
	if prompt != "" {
		args = append(args, fmt.Sprintf("--prompt=%s", prompt))
	}
	if len(expects) > 0 {
		args = append(args, fmt.Sprintf("--expect=%s", strings.Join(expects, ",")))
	}

	cmd := exec.Command("fzf", args...)
	cmd.Stderr = os.Stderr

	// Pipe items to stdin
	var buf bytes.Buffer
	for _, item := range items {
		buf.WriteString(item + "\n")
	}
	cmd.Stdin = &buf

	output, err := cmd.Output()
	if err != nil {
		// If fzf exits with non-zero (e.g. cancelled), we return an error
		return "", "", fmt.Errorf("selection cancelled or failed: %w", err)
	}

	// Parse output
	// When --expect is used, fzf outputs:
	// Line 1: Key pressed (empty if enter)
	// Line 2: Selected item
	if len(expects) > 0 {
		lines := strings.SplitN(string(output), "\n", 2)
		var key, val string
		if len(lines) >= 1 {
			key = strings.TrimSpace(lines[0])
		}
		if len(lines) >= 2 {
			val = strings.TrimSpace(lines[1])
		}
		return val, key, nil
	}

	return strings.TrimSpace(string(output)), "", nil
}
