package ui

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Select opens fzf with the provided items and returns the selected item.
func Select(items []string) (string, error) {
	cmd := exec.Command("fzf", "--ansi", "--no-sort", "--layout=reverse")
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
