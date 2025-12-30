package ui

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var ErrCancelled = errors.New("selection cancelled")

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

	// Map expected keys to --expect flag
	if len(expects) > 0 {
		args = append(args, "--expect="+strings.Join(expects, ","))
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
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 130 {
			return "", "", ErrCancelled
		}
		return "", "", fmt.Errorf("selection cancelled or failed: %w", err)
	}

	outStr := string(output)

	// If --expect was used, fzf prints the key on the first line and the selection on the second
	if len(expects) > 0 {
		lines := strings.SplitN(outStr, "\n", 2)
		if len(lines) >= 1 {
			key := strings.TrimSpace(lines[0])
			selection := ""
			if len(lines) == 2 {
				selection = strings.TrimSpace(lines[1])
			}
			return selection, key, nil
		}
	}

	return strings.TrimSpace(outStr), "", nil
}
