package ui

import (
	"fmt"
	"io"

	"atelier-go/internal/config"
	"atelier-go/internal/locations"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LocationItem wraps locations.Location for list display.
type LocationItem struct {
	Location locations.Location
}

// Title returns the formatted name of the location with an icon.
func (i LocationItem) Title() string {
	icon := IconFolder
	if i.Location.Source == "Project" {
		icon = IconProject
	}
	return fmt.Sprintf("%s %s", icon, i.Location.Name)
}

// Description returns the filesystem path of the location.
func (i LocationItem) Description() string { return i.Location.Path }

// FilterValue returns the string used for filtering locations.
func (i LocationItem) FilterValue() string { return i.Location.Name }

// IsProject returns true if the location is a project.
func (i LocationItem) IsProject() bool {
	return i.Location.Source == "Project"
}

// HasActions returns true if the location has associated actions.
func (i LocationItem) HasActions() bool {
	return len(i.Location.Actions) > 0
}

// ActionItem wraps config.Action for list display.
type ActionItem struct {
	Action    config.Action
	IsDefault bool
}

// Title returns the formatted name of the action.
func (a ActionItem) Title() string {
	if a.IsDefault {
		return a.Action.Name + " (Default)"
	}
	return a.Action.Name
}

// Description returns an empty string for action items.
func (a ActionItem) Description() string { return "" }

// FilterValue returns the string used for filtering actions.
func (a ActionItem) FilterValue() string { return a.Action.Name }

// LocationDelegate renders location items with focus-aware styling.
type LocationDelegate struct {
	NormalStyle   lipgloss.Style
	SelectedStyle lipgloss.Style
	Focused       bool
}

// NewLocationDelegate creates a new LocationDelegate with default styling.
func NewLocationDelegate(styles Styles) LocationDelegate {
	return LocationDelegate{
		NormalStyle:   styles.DelegateNormal,
		SelectedStyle: styles.DelegateSelected,
		Focused:       true,
	}
}

// Height returns the number of lines a single item occupies.
func (d LocationDelegate) Height() int { return 1 }

// Spacing returns the vertical spacing between items.
func (d LocationDelegate) Spacing() int { return 0 }

// Update handles logic for delegate updates.
func (d LocationDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

// Render paints the location item to the terminal.
func (d LocationDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(LocationItem)
	if !ok {
		return
	}

	icon := IconFolder
	if item.IsProject() {
		icon = IconProject
	}

	if index == m.Index() {
		style := d.SelectedStyle
		if !d.Focused {
			style = style.Foreground(ColorSubtext).BorderForeground(ColorSubtext)
		}
		_, _ = fmt.Fprint(w, style.Render(fmt.Sprintf("%s %s", icon, item.Location.Name)))
	} else {
		iconStyle := lipgloss.NewStyle().Foreground(ColorSubtext)
		textStyle := d.NormalStyle.Foreground(ColorText)

		if item.IsProject() {
			iconStyle = iconStyle.Foreground(ColorAccent)
			textStyle = textStyle.Foreground(ColorAccent)
		}

		_, _ = fmt.Fprint(w, iconStyle.Render(icon)+" "+textStyle.Render(item.Location.Name))
	}
}

// ActionDelegate renders action items with focus-aware styling.
type ActionDelegate struct {
	NormalStyle   lipgloss.Style
	SelectedStyle lipgloss.Style
	Focused       bool
}

// NewActionDelegate creates a new ActionDelegate with default styling.
func NewActionDelegate(styles Styles) ActionDelegate {
	return ActionDelegate{
		NormalStyle:   styles.DelegateNormal,
		SelectedStyle: styles.DelegateSelected,
		Focused:       false,
	}
}

// Height returns the number of lines a single item occupies.
func (d ActionDelegate) Height() int { return 1 }

// Spacing returns the vertical spacing between items.
func (d ActionDelegate) Spacing() int { return 0 }

// Update handles logic for delegate updates.
func (d ActionDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

// Render paints the action item to the terminal.
func (d ActionDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(ActionItem)
	if !ok {
		return
	}

	var style lipgloss.Style
	if index == m.Index() {
		style = d.SelectedStyle
		if !d.Focused {
			style = style.Foreground(ColorSubtext).BorderForeground(ColorSubtext)
		}
	} else {
		style = d.NormalStyle.Foreground(ColorText)
	}

	_, _ = fmt.Fprint(w, style.Render(item.Title()))
}
