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
	// Added --height=40% to match legacy behavior and potentially fix cursor issues.
	// Switched from --expect to --bind to avoid potential Esc key handling issues.
	args := []string{
		"--ansi",
		"--no-sort",
		"--layout=reverse",
		"--height=40%",
		"--bind=esc:abort",
	}

	if header != "" {
		args = append(args, fmt.Sprintf("--header=%s", header))
	}
	if prompt != "" {
		args = append(args, fmt.Sprintf("--prompt=%s", prompt))
	}

	// Map expected keys to bindings that print the key and accept
	for _, key := range expects {
		// print(key) writes to stdout (same stream as accept)
		// We use a null byte as a delimiter to safely separate key and selection
		args = append(args, fmt.Sprintf("--bind=%s:print(%s)+accept", key, key))
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

	outStr := strings.TrimSpace(string(output))

	// Check if any expected key prefixes the output
	for _, key := range expects {
		if strings.HasPrefix(outStr, key) {
			// Found a bound key
			val := strings.TrimPrefix(outStr, key)
			return strings.TrimSpace(val), key, nil
		}
	}

	return outStr, "", nil
}
