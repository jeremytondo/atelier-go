package ui

import (
	"atelier-go/internal/locations"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sahilm/fuzzy"
)

func (m *Model) syncFilter() []tea.Cmd {
	var cmds []tea.Cmd
	val := m.filterInput.Value()

	if m.focus == FocusLocations {
		if val != m.lastLeftFilter {
			cmds = append(cmds, m.applyLocationFilter())
			m.lastLeftFilter = val
			cmds = append(cmds, m.updateActions())
		}
	} else {
		if val != m.lastRightFilter {
			m.actions.SetFilterText(val)
			m.lastRightFilter = val
			m.actions.Select(0)
		}
	}
	return cmds
}

// locationSource implements fuzzy.Source for location matching.
type locationSource []locations.Location

func (s locationSource) String(i int) string { return s[i].Name }
func (s locationSource) Len() int            { return len(s) }

func (m *Model) applyLocationFilter() tea.Cmd {
	val := m.filterInput.Value()
	var items []list.Item

	if val == "" {
		for _, loc := range m.allLocations {
			items = append(items, LocationItem{Location: loc})
		}
	} else {
		matches := fuzzy.FindFrom(val, locationSource(m.allLocations))
		for _, match := range matches {
			items = append(items, LocationItem{Location: m.allLocations[match.Index]})
		}
	}

	m.locations.Select(0)
	return m.locations.SetItems(items)
}
