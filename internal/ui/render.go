package ui

import (
	"atelier-go/internal/locations"
	"atelier-go/internal/utils"
	"fmt"
	"regexp"
)

const (
	projectColor = "\x1b[38;2;137;180;250m"
	dimColor     = "\x1b[2m"
	resetColor   = "\x1b[0m"
	iconProject  = "\uf503"
	iconZoxide   = "\uea83"
)

// FormatLocation returns a colored string representation of a location for the UI.
func FormatLocation(loc locations.Location) string {
	if loc.Source == "Project" {
		return fmt.Sprintf("%s%s %s%s", projectColor, iconProject, loc.Name, resetColor)
	}
	// Zoxide entries
	shortPath := utils.ShortenPath(loc.Path)
	return fmt.Sprintf("%s %s  %s(%s)%s", iconZoxide, loc.Name, dimColor, shortPath, resetColor)
}

// StripANSI removes ANSI color codes from a string.
func StripANSI(str string) string {
	const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
	var re = regexp.MustCompile(ansi)
	return re.ReplaceAllString(str, "")
}
