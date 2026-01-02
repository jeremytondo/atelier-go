package ui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

// Colors
var (
	ColorPrimary = lipgloss.Color("#89b4fa") // Project blue
	ColorAccent  = lipgloss.Color("170")     // Pink selection
	ColorBorder  = lipgloss.Color("62")      // Border
	ColorDimmed  = lipgloss.Color("240")     // Unfocused
)

// Icons (Nerd Font) - can be overridden by init()
var (
	IconFolder  = "\uea83"
	IconProject = "\uf503"
	IconSearch  = "\uf002"
)

func init() {
	if os.Getenv("NO_NERD_FONTS") != "" {
		IconFolder = "F"
		IconProject = "P"
		IconSearch = "S"
	}

	// Initialize with reasonable defaults
	InitStyles(100, 30)
}

// Styles - initialized with defaults, updated on resize
var (
	WindowStyle       lipgloss.Style
	LeftPanelStyle    lipgloss.Style
	RightPanelStyle   lipgloss.Style
	SearchInputStyle  lipgloss.Style
	FocusedTitleStyle lipgloss.Style
	HelpStyle         lipgloss.Style
)

// Dimensions - updated on resize
var (
	ContentWidth int
	LeftWidth    int
	RightWidth   int
	ListHeight   int
)

// InitStyles recalculates styles based on terminal size
func InitStyles(termWidth, termHeight int) {
	// Constrain content width: min 60, max 120
	ContentWidth = termWidth - 4
	if ContentWidth > 120 {
		ContentWidth = 120
	}
	if ContentWidth < 60 {
		ContentWidth = 60
	}

	// Panel widths: 60/40 split
	LeftWidth = ContentWidth * 60 / 100
	RightWidth = ContentWidth - LeftWidth - 3 // Account for border

	// List height: leave room for search + borders
	ListHeight = termHeight - 8
	if ListHeight < 5 {
		ListHeight = 5
	}
	if ListHeight > 20 {
		ListHeight = 20
	}

	WindowStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(0, 1)

	LeftPanelStyle = lipgloss.NewStyle().
		Width(LeftWidth).
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(ColorDimmed)

	RightPanelStyle = lipgloss.NewStyle().
		Width(RightWidth).
		PaddingLeft(2)

	SearchInputStyle = lipgloss.NewStyle().
		Padding(1, 0)

	FocusedTitleStyle = lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true)

	HelpStyle = lipgloss.NewStyle().
		Foreground(ColorDimmed).
		MarginTop(1)
}
