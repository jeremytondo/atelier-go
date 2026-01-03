package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleEscape() []tea.Cmd {
	var cmds []tea.Cmd
	if m.focus == FocusActions {
		m.setFocus(FocusLocations)
		m.filterInput.SetValue(m.lastLeftFilter)
		m.actions.SetFilterText("")
		m.lastRightFilter = ""
		cmds = append(cmds, m.applyLocationFilter())
		cmds = append(cmds, m.updateActions())
		return cmds
	}

	if m.filterInput.Value() == "" {
		m.Result = SelectionResult{Canceled: true}
		m.quitting = true
		return nil
	}

	// Clear filter
	m.filterInput.SetValue("")
	cmds = append(cmds, m.applyLocationFilter())
	m.lastLeftFilter = ""
	cmds = append(cmds, m.updateActions())
	return cmds
}

func (m *Model) handleSelect() []tea.Cmd {
	var cmds []tea.Cmd
	if m.focus == FocusLocations {
		sel, ok := m.locations.SelectedItem().(LocationItem)
		if !ok {
			return nil
		}

		if sel.HasActions() {
			// Drill into actions
			m.setFocus(FocusActions)
			m.lastLeftFilter = m.filterInput.Value()
			m.filterInput.SetValue("")
			m.actions.Select(0)
			cmds = append(cmds, m.updateActions())
			return cmds
		}

		// No actions - select with default
		loc := sel.Location
		m.Result = SelectionResult{Location: &loc, Action: nil}
		m.quitting = true
		return nil
	}

	// In actions panel
	locItem, _ := m.locations.SelectedItem().(LocationItem)
	actItem, ok := m.actions.SelectedItem().(ActionItem)
	if !ok {
		return nil
	}

	loc := locItem.Location
	act := actItem.Action
	m.Result = SelectionResult{Location: &loc, Action: &act}
	m.quitting = true
	return nil
}

func (m *Model) handleFastSelect() {
	sel, ok := m.locations.SelectedItem().(LocationItem)
	if !ok {
		return
	}

	loc := sel.Location
	m.Result = SelectionResult{Location: &loc, Action: nil}
	m.quitting = true
}

func (m *Model) handleCursorUp() tea.Cmd {
	if m.focus == FocusLocations {
		m.locations.CursorUp()
		return m.updateActions()
	}
	m.actions.CursorUp()
	return nil
}

func (m *Model) handleCursorDown() tea.Cmd {
	if m.focus == FocusLocations {
		m.locations.CursorDown()
		return m.updateActions()
	}
	m.actions.CursorDown()
	return nil
}
