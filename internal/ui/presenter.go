package ui

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"atelier-go/internal/locations"
)

const (
	projectColor = "\x1b[38;2;137;180;250m"
	dimColor     = "\x1b[2m"
	resetColor   = "\x1b[0m"
	iconProject  = "\uf503"
	iconZoxide   = "\uea83"
)

// FormatItem returns a colored string representation of an item for the UI.
func FormatItem(item locations.Item) string {
	if item.Source == "Project" {
		return fmt.Sprintf("%s%s %s%s", projectColor, iconProject, item.Name, resetColor)
	}
	// Zoxide entries
	shortPath := prettyPath(item.Path)
	return fmt.Sprintf("%s %s  %s(%s)%s", iconZoxide, item.Name, dimColor, shortPath, resetColor)
}

// StripANSI removes ANSI color codes from a string.
func StripANSI(str string) string {
	const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
	var re = regexp.MustCompile(ansi)
	return re.ReplaceAllString(str, "")
}

func prettyPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if after, found := strings.CutPrefix(path, home); found {
		return "~" + after
	}
	return path
}
