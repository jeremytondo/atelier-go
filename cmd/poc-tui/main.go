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

var (
	titleStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1)
	windowStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1).
			Width(60)
	projectBlue = lipgloss.Color("#89b4fa")
)

type item struct {
	title     string
	isProject bool
}

func (i item) Title() string {
	icon := "\uea83"
	if i.isProject {
		icon = "\uf503"
	}
	return fmt.Sprintf("%s %s", icon, i.title)
}
func (i item) Description() string { return "" }
func (i item) FilterValue() string { return i.title }

type itemDelegate struct {
	list.DefaultDelegate
}

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	var style lipgloss.Style
	if index == m.Index() {
		style = d.Styles.SelectedTitle
	} else {
		style = d.Styles.NormalTitle
		if i.isProject {
			style = style.Copy().Foreground(projectBlue)
		}
	}

	fmt.Fprintf(w, style.Render(i.Title()))
}

type state int

const (
	stateProjects state = iota
	stateActions
)

type model struct {
	list           list.Model
	filterInput    textinput.Model
	choice         string
	action         string
	key            string
	quitting       bool
	prevFilter     string
	state          state
	project        string
	terminalWidth  int
	terminalHeight int
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.state == stateActions {
				m.state = stateProjects
				m.filterInput.SetValue("")
				m.list.SetFilterText("")
				m.prevFilter = ""
				m.filterInput.Prompt = "Project ➜ "
				cmd := m.list.SetItems(projectItems)
				m.list.Select(0)
				return m, cmd
			}
			if m.filterInput.Value() == "" {
				m.quitting = true
				return m, tea.Quit
			}
			m.filterInput.SetValue("")
			m.list.SetFilterText("")
			m.prevFilter = ""
			m.list.Select(0)
			return m, nil
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if !ok {
				return m, nil
			}
			if m.state == stateProjects {
				m.choice = i.title
				m.key = "enter"
				return m, tea.Quit
			} else {
				m.action = i.title
				return m, tea.Quit
			}
		case "alt+enter":
			if m.state == stateProjects {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					m.project = i.title
					m.state = stateActions
					m.filterInput.SetValue("")
					m.list.SetFilterText("")
					m.prevFilter = ""
					m.filterInput.Prompt = "Action ➜ "
					cmd := m.list.SetItems(actionItems)
					m.list.Select(0)
					return m, cmd
				}
			}
		case "up", "ctrl+p":
			m.list.CursorUp()
			return m, nil
		case "down", "ctrl+n":
			m.list.CursorDown()
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
		m.list.SetSize(60, 10)
	}

	m.filterInput, cmd = m.filterInput.Update(msg)
	cmds = append(cmds, cmd)

	newFilter := m.filterInput.Value()
	if newFilter != m.prevFilter {
		m.list.SetFilterText(newFilter)
		m.prevFilter = newFilter
		m.list.Select(0)
	}

	if _, ok := msg.(tea.KeyMsg); !ok {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	title := "Atelier Go Unified"
	if m.state == stateActions {
		title = fmt.Sprintf("Actions for %s", m.project)
	}
	header := titleStyle.Render(title)
	search := "\n" + m.filterInput.View() + "\n"

	inner := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		search,
		m.list.View(),
	)

	content := windowStyle.Render(inner)

	if m.terminalWidth == 0 {
		return content
	}

	return lipgloss.Place(m.terminalWidth, m.terminalHeight,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

var (
	projectItems = []list.Item{
		item{title: "Atelier Go", isProject: true},
		item{title: "Bubble Tea", isProject: true},
		item{title: "Charm", isProject: false},
		item{title: "Fzf Replacement", isProject: false},
		item{title: "Golang Project", isProject: false},
		item{title: "TUI Exploration", isProject: false},
	}
	actionItems = []list.Item{
		item{title: "Default (Shell)", isProject: false},
		item{title: "Open in Editor", isProject: false},
		item{title: "Run Tests", isProject: false},
		item{title: "Build Project", isProject: false},
	}
)

func main() {
	defaultDelegate := list.NewDefaultDelegate()
	defaultDelegate.ShowDescription = false
	defaultDelegate.SetSpacing(0)
	defaultDelegate.Styles.NormalTitle = lipgloss.NewStyle().Padding(0, 0, 0, 1)
	defaultDelegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderLeftForeground(lipgloss.Color("170")).
		Foreground(lipgloss.Color("170")).
		Padding(0, 0, 0, 1)

	delegate := itemDelegate{defaultDelegate}

	l := list.New(projectItems, delegate, 60, 10)
	l.SetShowTitle(false)
	l.SetShowFilter(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)

	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Focus()
	ti.Prompt = "Project ➜ "
	ti.CharLimit = 64
	ti.Width = 30

	m := model{
		list:        l,
		filterInput: ti,
		state:       stateProjects,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := finalModel.(model); ok {
		if m.action != "" {
			fmt.Printf("\nProject: %s\nAction: %s\n", m.project, m.action)
		} else if m.choice != "" {
			fmt.Printf("\nSelection: %s\n", m.choice)
		} else {
			fmt.Println("\nCancelled")
		}
	}
}
