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

// Icons (Nerd Font)
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
}

// Layout defines the dimensions and proportions of the TUI panels.
type Layout struct {
	Width        int
	Height       int
	ContentWidth int
	LeftWidth    int
	RightWidth   int
	ListHeight   int
}

// Styles holds all the lipgloss styles used in the TUI.
type Styles struct {
	Window       lipgloss.Style
	LeftPanel    lipgloss.Style
	RightPanel   lipgloss.Style
	SearchInput  lipgloss.Style
	FocusedTitle lipgloss.Style
	Help         lipgloss.Style
}

// DefaultLayout returns a Layout based on the provided terminal dimensions.
func DefaultLayout(termWidth, termHeight int) Layout {
	// Constrain content width: min 60, max 120
	contentWidth := termWidth - 4
	if contentWidth > 120 {
		contentWidth = 120
	}
	if contentWidth < 60 {
		contentWidth = 60
	}

	// Panel widths: 60/40 split
	leftWidth := contentWidth * 60 / 100
	rightWidth := contentWidth - leftWidth - 3 // Account for border

	// List height: leave room for search + borders
	listHeight := termHeight - 8
	if listHeight < 5 {
		listHeight = 5
	}
	if listHeight > 20 {
		listHeight = 20
	}

	return Layout{
		Width:        termWidth,
		Height:       termHeight,
		ContentWidth: contentWidth,
		LeftWidth:    leftWidth,
		RightWidth:   rightWidth,
		ListHeight:   listHeight,
	}
}

// DefaultStyles returns the default Styles based on a Layout.
func DefaultStyles(l Layout) Styles {
	return Styles{
		Window: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1),

		LeftPanel: lipgloss.NewStyle().
			Width(l.LeftWidth).
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(ColorDimmed),

		RightPanel: lipgloss.NewStyle().
			Width(l.RightWidth).
			PaddingLeft(2),

		SearchInput: lipgloss.NewStyle().
			Padding(1, 0),

		FocusedTitle: lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true),

		Help: lipgloss.NewStyle().
			Foreground(ColorDimmed).
			MarginTop(1),
	}
}
