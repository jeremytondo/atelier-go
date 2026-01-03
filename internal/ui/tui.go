package ui

import (
	"atelier-go/internal/config"
	"atelier-go/internal/locations"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
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

	// Initial layout and styles (will be updated on first resize)
	layout := DefaultLayout(100, 30)
	styles := DefaultStyles(layout)

	// Delegates
	locDelegate := NewLocationDelegate()
	actDelegate := NewActionDelegate()

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
		// Add Shell option for projects with actions
		if sel.IsProject() && len(sel.Location.Actions) > 0 {
			items = append(items, ActionItem{
				Action: config.Action{Name: "Shell", Command: ""},
			})
		}
	}

	if len(items) == 0 {
		return m.actions.SetItems(nil)
	}

	return m.actions.SetItems(items)
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

func (m *Model) updateDimensions() {
	m.locations.SetSize(m.layout.LeftWidth, m.layout.ListHeight)
	m.actions.SetSize(m.layout.RightWidth, m.layout.ListHeight)
	m.filterInput.Width = m.layout.ContentWidth - 10
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

	help := m.styles.Help.Render("Enter:Select • Tab:Actions • Alt+Enter:Fast • Esc:Back • Ctrl+C:Quit")

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
