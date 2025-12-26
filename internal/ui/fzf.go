package ui

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Select opens fzf with the provided items and returns the selected item.
func Select(items []string, header string, prompt string) (string, error) {
	args := []string{"--ansi", "--no-sort", "--layout=reverse"}
	if header != "" {
		args = append(args, fmt.Sprintf("--header=%s", header))
	}
	if prompt != "" {
		args = append(args, fmt.Sprintf("--prompt=%s", prompt))
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
		return "", fmt.Errorf("selection cancelled or failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
