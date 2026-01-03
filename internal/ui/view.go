package ui

import (
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) updateDimensions() {
	m.locations.SetSize(m.layout.LeftWidth, m.layout.ListHeight)
	m.actions.SetSize(m.layout.RightWidth, m.layout.ListHeight)
	m.filterInput.Width = m.layout.ContentWidth - 10

	// Update delegate styles
	m.locationsDelegate.NormalStyle = m.styles.DelegateNormal
	m.locationsDelegate.SelectedStyle = m.styles.DelegateSelected
	m.locations.SetDelegate(m.locationsDelegate)

	m.actionsDelegate.NormalStyle = m.styles.DelegateNormal
	m.actionsDelegate.SelectedStyle = m.styles.DelegateSelected
	m.actions.SetDelegate(m.actionsDelegate)
}

// View renders the TUI to the terminal.
func (m *Model) View() string {
	if m.quitting {
		return ""
	}

	search := m.styles.SearchInput.Render(m.filterInput.View())

	locationTitle := m.styles.NormalTitle.Render("PROJECTS & LOCATIONS")
	if m.focus == FocusLocations {
		locationTitle = m.styles.FocusedTitle.Render("SELECT LOCATION")
	}

	var leftView string
	if len(m.locations.Items()) == 0 {
		leftView = m.styles.LeftPanel.Render(
			lipgloss.JoinVertical(lipgloss.Left, locationTitle, "", "No items match your search"),
		)
	} else {
		leftView = m.styles.LeftPanel.Render(
			lipgloss.JoinVertical(lipgloss.Left, locationTitle, "", m.locations.View()),
		)
	}

	actionTitle := m.styles.NormalTitle.Render("AVAILABLE ACTIONS")
	if m.focus == FocusActions {
		actionTitle = m.styles.FocusedTitle.Render("SELECT ACTION")
	}

	var rightView string
	if len(m.actions.Items()) == 0 {
		rightView = "No actions available"
	} else {
		rightView = m.actions.View()
	}

	right := m.styles.RightPanel.Render(
		lipgloss.JoinVertical(lipgloss.Left, actionTitle, "", rightView),
	)

	help := m.styles.Help.Render("Enter:Select • Tab:Actions • Alt+Enter:Default Action • Esc:Back • Ctrl+C:Quit")

	inner := lipgloss.JoinVertical(
		lipgloss.Left,
		search,
		lipgloss.JoinHorizontal(lipgloss.Top, leftView, right),
		help,
	)

	content := m.styles.Window.Render(inner)

	if m.layout.Width == 0 {
		return content
	}

	return lipgloss.Place(m.layout.Width, m.layout.Height, lipgloss.Center, lipgloss.Center, content)
}
