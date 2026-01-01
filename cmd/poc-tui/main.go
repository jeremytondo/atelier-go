package main

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Visual constants for the TUI
var (
	windowStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1).
			Width(100)
	projectBlue = lipgloss.Color("#89b4fa")

	// Panel layout styles
	leftPanelStyle = lipgloss.NewStyle().
			Width(60).
			Border(lipgloss.NormalBorder(), false, true, false, false). // Vertical divider
			BorderForeground(lipgloss.Color("240"))

	rightPanelStyle = lipgloss.NewStyle().
			Width(35).
			PaddingLeft(2)

	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))

	searchInputStyle = lipgloss.NewStyle().
				Bold(true).
				MarginBottom(1)
)

// item represents a single selectable entry in our lists (Project, Folder, or Action)
type item struct {
	title     string
	isProject bool
	actions   []list.Item
}

func (i item) Title() string {
	icon := "\uea83" // Folder icon
	if i.isProject {
		icon = "\uf503" // Project icon
	}
	// Return the formatted title with icon
	return fmt.Sprintf("%s %s", icon, i.title)
}
func (i item) Description() string { return "" }
func (i item) FilterValue() string { return i.title }

// itemDelegate handles rendering of list items with custom focusing logic
type itemDelegate struct {
	list.DefaultDelegate
	focused bool // Controls whether the selection indicator is highlighted
}

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	var style lipgloss.Style
	if index == m.Index() {
		style = d.Styles.SelectedTitle
		if !d.focused {
			// Dim the selection color if the panel is not currently focused
			style = style.Copy().Foreground(lipgloss.Color("240")).BorderLeftForeground(lipgloss.Color("240"))
		}
	} else {
		style = d.Styles.NormalTitle
		if i.isProject {
			style = style.Copy().Foreground(projectBlue)
		}
	}

	fmt.Fprintf(w, style.Render(i.Title()))
}

type focus int

const (
	focusLeft focus = iota
	focusRight
)

// model holds the state of the entire TUI application
type model struct {
	projects         list.Model
	projectsDelegate itemDelegate
	actions          list.Model
	actionsDelegate  itemDelegate
	filterInput      textinput.Model
	focus            focus
	terminalWidth    int
	terminalHeight   int
	quitting         bool

	// Final results to be printed after exit
	SelectedProject string
	SelectedAction  string

	// Track filter states for both panels
	lastLeftFilter  string
	lastRightFilter string
}

func (m *model) Init() tea.Cmd {
	// Initialize with cursor blinking and ensure the first item's actions are loaded
	return tea.Batch(textinput.Blink, m.updateActions())
}

// updateActions synchronizes the right panel with the current selection on the left
func (m *model) updateActions() tea.Cmd {
	var items []list.Item
	if sel, ok := m.projects.SelectedItem().(item); ok {
		if len(sel.actions) > 0 {
			items = sel.actions
		} else {
			items = []list.Item{item{title: "No actions available"}}
		}
	}
	return m.actions.SetItems(items)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "esc":
			if m.focus == focusRight {
				// Go back to project list
				m.focus = focusLeft
				m.filterInput.SetValue(m.lastLeftFilter)
				m.filterInput.Prompt = "Project ➜ "
				m.actions.SetFilterText("")
				m.lastRightFilter = ""
				cmds = append(cmds, m.updateActions())
				return m, tea.Batch(cmds...)
			}
			if m.filterInput.Value() == "" {
				m.quitting = true
				return m, tea.Quit
			}
			// Clear filter
			m.filterInput.SetValue("")
			m.projects.SetFilterText("")
			m.lastLeftFilter = ""
			cmds = append(cmds, m.updateActions())
			return m, tea.Batch(cmds...)

		case "enter", "tab":
			if m.focus == focusLeft {
				if sel, ok := m.projects.SelectedItem().(item); ok && sel.isProject {
					// Switch to actions panel
					m.focus = focusRight
					m.lastLeftFilter = m.filterInput.Value()
					m.filterInput.SetValue("")
					m.filterInput.Prompt = "Action ➜ "
					m.actions.Select(0)
					return m, nil
				}
				// Folder or non-project selected -> Execute default
				if sel, ok := m.projects.SelectedItem().(item); ok {
					m.SelectedProject = sel.title
					m.SelectedAction = "Default"
				}
				m.quitting = true
				return m, tea.Quit
			} else {
				// In actions panel, Enter selects the action
				if p, ok := m.projects.SelectedItem().(item); ok {
					m.SelectedProject = p.title
				}
				if a, ok := m.actions.SelectedItem().(item); ok {
					m.SelectedAction = a.title
				}
				m.quitting = true
				return m, tea.Quit
			}

		case "alt+enter", "ctrl+s":
			// "Fast" trigger: bypass action menu and use default.
			// Note: ctrl+enter is omitted as it's unreliable in most terminal emulators.
			if p, ok := m.projects.SelectedItem().(item); ok {
				m.SelectedProject = p.title
				m.SelectedAction = "Default"
			}
			m.quitting = true
			return m, tea.Quit

		case "up", "ctrl+p":
			if m.focus == focusLeft {
				m.projects.CursorUp()
				cmds = append(cmds, m.updateActions())
			} else {
				m.actions.CursorUp()
			}
			return m, tea.Batch(cmds...)

		case "down", "ctrl+n":
			if m.focus == focusLeft {
				m.projects.CursorDown()
				cmds = append(cmds, m.updateActions())
			} else {
				m.actions.CursorDown()
			}
			return m, tea.Batch(cmds...)
		}

	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
	}

	// Update Filter Input
	m.filterInput, cmd = m.filterInput.Update(msg)
	cmds = append(cmds, cmd)

	// Synchronize Filters between Textinput and the active List
	if m.focus == focusLeft {
		val := m.filterInput.Value()
		if val != m.lastLeftFilter {
			m.projects.SetFilterText(val)
			m.lastLeftFilter = val
			m.projects.Select(0)
			cmds = append(cmds, m.updateActions())
		}
	} else {
		val := m.filterInput.Value()
		if val != m.lastRightFilter {
			m.actions.SetFilterText(val)
			m.lastRightFilter = val
			m.actions.Select(0)
		}
	}

	// Route non-key messages to both lists to keep them reactive
	if _, isKey := msg.(tea.KeyMsg); !isKey {
		m.projects, cmd = m.projects.Update(msg)
		cmds = append(cmds, cmd)
		m.actions, cmd = m.actions.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) View() string {
	if m.quitting {
		return ""
	}

	// Update delegates with current focus state before rendering
	m.projectsDelegate.focused = (m.focus == focusLeft)
	m.projects.SetDelegate(m.projectsDelegate)

	m.actionsDelegate.focused = (m.focus == focusRight)
	m.actions.SetDelegate(m.actionsDelegate)

	search := searchInputStyle.Render(m.filterInput.View())

	// Panels
	left := leftPanelStyle.Render(m.projects.View())

	// Right panel title styling based on focus
	actionTitle := "AVAILABLE ACTIONS"
	if m.focus == focusRight {
		actionTitle = focusedStyle.Bold(true).Render("SELECT ACTION")
	}

	right := rightPanelStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			actionTitle,
			"",
			m.actions.View(),
		),
	)

	// Compose the layout
	inner := lipgloss.JoinVertical(
		lipgloss.Left,
		search,
		lipgloss.JoinHorizontal(lipgloss.Top, left, right),
	)

	content := windowStyle.Render(inner)

	if m.terminalWidth == 0 {
		return content
	}

	// Center the window in the terminal
	return lipgloss.Place(m.terminalWidth, m.terminalHeight,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

func main() {
	// Sample action template
	defaultActions := []list.Item{
		item{title: "Shell", isProject: false},
		item{title: "Editor", isProject: false},
		item{title: "Test", isProject: false},
		item{title: "Build", isProject: false},
	}

	// Sample project data
	projectItems := []list.Item{
		item{title: "atelier-go", isProject: true, actions: defaultActions},
		item{title: "bubbletea", isProject: true, actions: defaultActions},
		item{title: "charm-bubbles", isProject: false},
		item{title: "fzf-core", isProject: false},
		item{title: "golang-poc", isProject: true, actions: defaultActions},
		item{title: "tui-exploration", isProject: false},
	}

	// Setup the shared delegate
	d := list.NewDefaultDelegate()
	d.ShowDescription = false
	d.SetSpacing(0)
	d.Styles.NormalTitle = lipgloss.NewStyle().Padding(0, 0, 0, 1)
	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderLeftForeground(lipgloss.Color("170")).
		Foreground(lipgloss.Color("170")).
		Padding(0, 0, 0, 1)

	delegate := itemDelegate{DefaultDelegate: d}

	// Initialize the panel lists
	l := list.New(projectItems, delegate, 58, 10)
	l.SetShowTitle(false)
	l.SetShowFilter(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)

	r := list.New(nil, delegate, 30, 10)
	r.SetShowTitle(false)
	r.SetShowFilter(false)
	r.SetShowStatusBar(false)
	r.SetShowHelp(false)

	// Initialize the shared filter input
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Focus()
	ti.Prompt = "Project ➜ "
	ti.CharLimit = 64
	ti.Width = 50
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Bold(true)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	m := &model{
		projects:         l,
		projectsDelegate: delegate,
		actions:          r,
		actionsDelegate:  delegate,
		filterInput:      ti,
		focus:            focusLeft,
	}

	// Start the Bubble Tea program
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// PRINT SELECTIONS AFTER EXIT
	if m, ok := finalModel.(*model); ok && m.SelectedProject != "" {
		fmt.Printf("\nSelection: %s\n", m.SelectedProject)
		if m.SelectedAction != "" {
			fmt.Printf("Action:    %s\n", m.SelectedAction)
		}
	} else {
		fmt.Println("\nSelection Cancelled")
	}
}
