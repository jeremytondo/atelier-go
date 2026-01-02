package ui

import (
	"fmt"
	"io"

	"atelier-go/internal/config"
	"atelier-go/internal/locations"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// LocationItem wraps locations.Location for list display
type LocationItem struct {
	Location locations.Location
}

func (i LocationItem) Title() string {
	icon := IconFolder
	if i.Location.Source == "Project" {
		icon = IconProject
	}
	return fmt.Sprintf("%s %s", icon, i.Location.Name)
}
func (i LocationItem) Description() string { return i.Location.Path }
func (i LocationItem) FilterValue() string { return i.Location.Name }

func (i LocationItem) IsProject() bool {
	return i.Location.Source == "Project"
}

func (i LocationItem) HasActions() bool {
	return len(i.Location.Actions) > 0
}

// ActionItem wraps config.Action for list display
type ActionItem struct {
	Action    config.Action
	IsDefault bool
}

func (a ActionItem) Title() string {
	if a.IsDefault {
		return a.Action.Name + " (Default)"
	}
	return a.Action.Name
}
func (a ActionItem) Description() string { return "" }
func (a ActionItem) FilterValue() string { return a.Action.Name }

// LocationDelegate renders location items with focus-aware styling
type LocationDelegate struct {
	list.DefaultDelegate
	Focused bool
}

func NewLocationDelegate() LocationDelegate {
	d := list.NewDefaultDelegate()
	d.ShowDescription = false
	d.SetSpacing(0)
	d.Styles.NormalTitle = lipgloss.NewStyle().Padding(0, 0, 0, 1)
	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(ColorAccent).
		Foreground(ColorAccent).
		Padding(0, 0, 0, 1)
	return LocationDelegate{DefaultDelegate: d, Focused: true}
}

func (d LocationDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(LocationItem)
	if !ok {
		return
	}

	var style lipgloss.Style
	if index == m.Index() {
		style = d.Styles.SelectedTitle
		if !d.Focused {
			style = style.Foreground(ColorDimmed).BorderForeground(ColorDimmed)
		}
	} else {
		style = d.Styles.NormalTitle
		if item.IsProject() {
			style = style.Foreground(ColorPrimary)
		}
	}

	fmt.Fprint(w, style.Render(item.Title()))
}

// ActionDelegate renders action items with focus-aware styling
type ActionDelegate struct {
	list.DefaultDelegate
	Focused bool
}

func NewActionDelegate() ActionDelegate {
	d := list.NewDefaultDelegate()
	d.ShowDescription = false
	d.SetSpacing(0)
	d.Styles.NormalTitle = lipgloss.NewStyle().Padding(0, 0, 0, 1)
	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(ColorAccent).
		Foreground(ColorAccent).
		Padding(0, 0, 0, 1)
	return ActionDelegate{DefaultDelegate: d, Focused: false}
}

func (d ActionDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(ActionItem)
	if !ok {
		return
	}

	var style lipgloss.Style
	if index == m.Index() {
		style = d.Styles.SelectedTitle
		if !d.Focused {
			style = style.Foreground(ColorDimmed).BorderForeground(ColorDimmed)
		}
	} else {
		style = d.Styles.NormalTitle
	}

	fmt.Fprint(w, style.Render(item.Title()))
}
