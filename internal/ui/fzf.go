// Package ui provides the interactive fuzzy search interface.
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
	args := []string{"--ansi", "--no-sort", "--layout=reverse"}
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

	outStr := strings.TrimSpace(string(output))
	if len(expects) > 0 {
		lines := strings.SplitN(outStr, "\n", 2)
		if len(lines) == 2 {
			return strings.TrimSpace(lines[1]), strings.TrimSpace(lines[0]), nil
		} else if len(lines) == 1 {
			// This might happen if nothing was selected but a key was pressed?
			// Or if just enter was pressed and it printed an empty line then the item?
			// If key is empty line, SplitN might give 2 lines if item is present.
			// Let's rely on line count.
			// Actually strings.TrimSpace above might remove the newline if key was empty.
			// Re-reading raw output is safer.
			// raw := string(output)
			// linesRaw := strings.Split(raw, "\n")
			// Filter out empty lines at the end usually
			// var validLines []string
			// for _, l := range linesRaw {
			// 	if l != "" {
			// 		validLines = append(validLines, l)
			// 	}
			// }

			// If raw output starts with newline, key is empty.
			key := ""
			val := ""

			// Based on fzf docs:
			// "The first line is the name of the key..."
			// "The second line is the selected line."

			// If I pressed Enter (and it wasn't expected), the key line is empty string.
			// So output is "\nSelection".

			lines = strings.SplitN(string(output), "\n", 2)
			if len(lines) >= 1 {
				key = strings.TrimSpace(lines[0])
			}
			if len(lines) >= 2 {
				val = strings.TrimSpace(lines[1])
			}
			return val, key, nil
		}
	}

	return outStr, "", nil
}
