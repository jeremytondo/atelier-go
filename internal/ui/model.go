package ui

import (
	"atelier-go/internal/config"
	"atelier-go/internal/locations"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Focus indicates which panel is active.
type Focus int

const (
	// FocusLocations represents the left panel (projects/directories).
	FocusLocations Focus = iota
	// FocusActions represents the right panel (available commands).
	FocusActions
)

// SelectionResult holds the final user selection.
type SelectionResult struct {
	Location *locations.Location
	Action   *config.Action
	Canceled bool
}

// Model is the Bubble Tea model for the TUI.
type Model struct {
	// Data
	allLocations []locations.Location

	// Components
	locations         list.Model
	locationsDelegate LocationDelegate
	actions           list.Model
	actionsDelegate   ActionDelegate
	filterInput       textinput.Model

	// State
	focus           Focus
	layout          Layout
	styles          Styles
	quitting        bool
	lastLeftFilter  string
	lastRightFilter string

	// Result
	Result SelectionResult
}

// NewModel creates a TUI model from locations.
func NewModel(locs []locations.Location) *Model {
	// Convert to list items
	items := make([]list.Item, len(locs))
	for i, loc := range locs {
		items[i] = LocationItem{Location: loc}
	}

	// Sort: Projects first, then maintain relative order
	sort.SliceStable(items, func(i, j int) bool {
		li := items[i].(LocationItem)
		lj := items[j].(LocationItem)
		if li.IsProject() && !lj.IsProject() {
			return true
		}
		if !li.IsProject() && lj.IsProject() {
			return false
		}
		return false
	})

	// Initial layout and styles (will be updated on first resize)
	layout := DefaultLayout(100, 30)
	styles := DefaultStyles(layout)

	// Delegates
	locDelegate := NewLocationDelegate(styles)
	actDelegate := NewActionDelegate(styles)

	// Location list
	locList := list.New(items, locDelegate, layout.LeftWidth, layout.ListHeight)
	locList.SetShowTitle(false)
	locList.SetShowFilter(false)
	locList.SetShowStatusBar(false)
	locList.SetShowHelp(false)
	locList.KeyMap.Filter.SetEnabled(false) // Disable internal filtering key

	// Action list (empty initially)
	actList := list.New(nil, actDelegate, layout.RightWidth, layout.ListHeight)
	actList.SetShowTitle(false)
	actList.SetShowFilter(false)
	actList.SetShowStatusBar(false)
	actList.SetShowHelp(false)

	// Filter input
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Focus()
	ti.Prompt = IconSearch + " "
	ti.CharLimit = 64
	ti.Width = layout.ContentWidth - 10
	ti.PromptStyle = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(ColorText).Bold(true)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(ColorSubtext)

	return &Model{
		allLocations:      locs,
		locations:         locList,
		locationsDelegate: locDelegate,
		actions:           actList,
		actionsDelegate:   actDelegate,
		filterInput:       ti,
		focus:             FocusLocations,
		layout:            layout,
		styles:            styles,
	}
}

// Init initializes the model.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.updateActions())
}

// Update handles terminal messages and user input.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// 1. Handle key messages
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c":
			m.Result = SelectionResult{Canceled: true}
			m.quitting = true
			return m, tea.Quit

		case "esc":
			cmds = append(cmds, m.handleEscape()...)
			if m.quitting {
				return m, tea.Quit
			}

		case "enter", "tab":
			cmds = append(cmds, m.handleSelect()...)
			if m.quitting {
				return m, tea.Quit
			}

		case "alt+enter", "ctrl+s":
			m.handleFastSelect()
			if m.quitting {
				return m, tea.Quit
			}

		case "up", "ctrl+p":
			cmds = append(cmds, m.handleCursorUp())

		case "down", "ctrl+n":
			cmds = append(cmds, m.handleCursorDown())
		}
	}

	// 2. Handle other messages
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.layout = DefaultLayout(msg.Width, msg.Height)
		m.styles = DefaultStyles(m.layout)
		m.updateDimensions()
	}

	// 3. Update components (always update filter input if not quitting)
	if !m.quitting {
		var cmd tea.Cmd
		m.filterInput, cmd = m.filterInput.Update(msg)
		cmds = append(cmds, cmd)

		// 4. Sync filter
		cmds = append(cmds, m.syncFilter()...)

		// 5. Route non-key messages to lists
		if _, isKey := msg.(tea.KeyMsg); !isKey {
			m.locations, cmd = m.locations.Update(msg)
			cmds = append(cmds, cmd)
			m.actions, cmd = m.actions.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// setFocus updates the active panel and its corresponding styling.
func (m *Model) setFocus(f Focus) {
	m.focus = f
	m.locationsDelegate.Focused = (m.focus == FocusLocations)
	m.locations.SetDelegate(m.locationsDelegate)
	m.actionsDelegate.Focused = (m.focus == FocusActions)
	m.actions.SetDelegate(m.actionsDelegate)

	// Update search prompt based on focus
	if f == FocusLocations {
		m.filterInput.Prompt = IconSearch + " "
	} else {
		m.filterInput.Prompt = "Action " + IconSearch + " "
	}
}

func (m *Model) updateActions() tea.Cmd {
	var items []list.Item

	if sel, ok := m.locations.SelectedItem().(LocationItem); ok {
		for i, act := range sel.Location.Actions {
			items = append(items, ActionItem{
				Action:    act,
				IsDefault: i == 0,
			})
		}
	}

	if len(items) == 0 {
		return m.actions.SetItems(nil)
	}

	return m.actions.SetItems(items)
}
