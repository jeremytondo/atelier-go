package ui

import (
	"atelier-go/internal/config"
	"os"

	"github.com/charmbracelet/lipgloss"
)

// Colors
var (
	ColorPrimary   = lipgloss.Color("#89b4fa") // Default border color
	ColorAccent    = lipgloss.Color("#74c7ec") // Default project/icon color
	ColorHighlight = lipgloss.Color("#cba6f7") // Default selection color
	ColorSubtext   = lipgloss.Color("240")     // Default subtext
	ColorText      = lipgloss.Color("#ffffff") // Default text color
)

// ApplyTheme overrides the default colors with values from the config.
func ApplyTheme(theme config.Theme) {
	if theme.Primary != "" {
		ColorPrimary = lipgloss.Color(theme.Primary)
	}
	if theme.Accent != "" {
		ColorAccent = lipgloss.Color(theme.Accent)
	}
	if theme.Highlight != "" {
		ColorHighlight = lipgloss.Color(theme.Highlight)
	}
	if theme.Text != "" {
		ColorText = lipgloss.Color(theme.Text)
	}
	if theme.Subtext != "" {
		ColorSubtext = lipgloss.Color(theme.Subtext)
	}
}

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
	Window           lipgloss.Style
	LeftPanel        lipgloss.Style
	RightPanel       lipgloss.Style
	SearchInput      lipgloss.Style
	FocusedTitle     lipgloss.Style
	NormalTitle      lipgloss.Style
	Help             lipgloss.Style
	DelegateNormal   lipgloss.Style
	DelegateSelected lipgloss.Style
}

// DefaultLayout returns a Layout based on the provided terminal dimensions.
func DefaultLayout(termWidth, termHeight int) Layout {
	// Constrain content width: min 60, max 120
	contentWidth := max(60, min(120, termWidth-4))

	// Panel widths: 60/40 split
	leftWidth := contentWidth * 60 / 100
	rightWidth := contentWidth - leftWidth - 3 // Account for border

	// List height: leave room for search + borders + titles
	listHeight := max(5, min(20, termHeight-11))

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
			BorderForeground(ColorPrimary).
			Padding(0, 1),

		LeftPanel: lipgloss.NewStyle().
			Width(l.LeftWidth).
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(ColorSubtext),

		RightPanel: lipgloss.NewStyle().
			Width(l.RightWidth).
			PaddingLeft(2),

		SearchInput: lipgloss.NewStyle().
			Width(l.ContentWidth-2). // Account for search box borders
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Foreground(ColorText).
			Padding(0, 1).
			MarginBottom(1),

		FocusedTitle: lipgloss.NewStyle().
			Foreground(ColorHighlight).
			Bold(true),

		NormalTitle: lipgloss.NewStyle().
			Foreground(ColorText).
			Bold(true),

		Help: lipgloss.NewStyle().
			Foreground(ColorSubtext).
			MarginTop(1),

		DelegateNormal: lipgloss.NewStyle().
			Padding(0, 0, 0, 1),

		DelegateSelected: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(ColorHighlight).
			Foreground(ColorHighlight).
			Padding(0, 0, 0, 1),
	}
}
