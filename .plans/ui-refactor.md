# UI Refactor: Native TUI with Bubble Tea

## Overview

Replace the `fzf` dependency with a native Go TUI using the Charm ecosystem. The POC at `cmd/poc-tui/main.go` validates the UX. This plan details the production implementation.

**Tech Stack:**
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - Components (list, textinput)
- `github.com/charmbracelet/lipgloss` - Styling

---

## Target File Structure

```
internal/ui/
  ui.go          # Public API - minimal changes, calls TUI
  tui.go         # NEW: Bubble Tea model, Update, View
  components.go  # NEW: LocationItem, ActionItem, delegates
  styles.go      # NEW: Lip Gloss style definitions
  render.go      # KEEP: FormatLocation, StripANSI utilities
  fzf.go         # DELETE after integration complete
```

---

## Data Types Reference

### Input Types (from existing packages)

```go
// locations.Location - what we receive from locations.Manager
type Location struct {
    Name    string          // Display name
    Path    string          // Filesystem path
    Source  string          // "Project" or "Zoxide"
    Actions []config.Action // Available actions
}

// config.Action - individual action definition
type Action struct {
    Name    string // Display name (e.g., "Editor", "Test")
    Command string // Shell command to execute
}
```

### Output Type (what TUI returns)

```go
// SelectionResult - returned from TUI to ui.Run()
type SelectionResult struct {
    Location *locations.Location // Selected location
    Action   *config.Action      // Selected action (nil = default)
    Canceled bool                // True if user pressed Esc/Ctrl+C
}
```

---

## Phase 1: Foundation

### Task 1.1: Create `internal/ui/styles.go`

**Purpose:** Centralize all visual styling with responsive width calculations.

```go
package ui

import "github.com/charmbracelet/lipgloss"

// Colors
var (
    ColorPrimary = lipgloss.Color("#89b4fa") // Project blue
    ColorAccent  = lipgloss.Color("170")     // Pink selection
    ColorBorder  = lipgloss.Color("62")      // Border
    ColorDimmed  = lipgloss.Color("240")     // Unfocused
)

// Icons (Nerd Font)
const (
    IconFolder  = "\uea83"
    IconProject = "\uf503"
    IconSearch  = "\uf002"
)

// Styles - initialized with defaults, updated on resize
var (
    WindowStyle      lipgloss.Style
    LeftPanelStyle   lipgloss.Style
    RightPanelStyle  lipgloss.Style
    SearchInputStyle lipgloss.Style
    FocusedTitleStyle lipgloss.Style
)

// Dimensions - updated on resize
var (
    ContentWidth int
    LeftWidth    int
    RightWidth   int
    ListHeight   int
)

// InitStyles recalculates styles based on terminal size
func InitStyles(termWidth, termHeight int) {
    // Constrain content width: min 60, max 120
    ContentWidth = termWidth - 4
    if ContentWidth > 120 {
        ContentWidth = 120
    }
    if ContentWidth < 60 {
        ContentWidth = 60
    }

    // Panel widths: 60/40 split
    LeftWidth = ContentWidth * 60 / 100
    RightWidth = ContentWidth - LeftWidth - 3 // Account for border

    // List height: leave room for search + borders
    ListHeight = termHeight - 8
    if ListHeight < 5 {
        ListHeight = 5
    }
    if ListHeight > 20 {
        ListHeight = 20
    }

    WindowStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(ColorBorder).
        Padding(0, 1)

    LeftPanelStyle = lipgloss.NewStyle().
        Width(LeftWidth).
        Border(lipgloss.NormalBorder(), false, true, false, false).
        BorderForeground(ColorDimmed)

    RightPanelStyle = lipgloss.NewStyle().
        Width(RightWidth).
        PaddingLeft(2)

    SearchInputStyle = lipgloss.NewStyle().
        Padding(1, 0)

    FocusedTitleStyle = lipgloss.NewStyle().
        Foreground(ColorAccent).
        Bold(true)
}

func init() {
    // Initialize with reasonable defaults
    InitStyles(100, 30)
}
```

### Task 1.2: Create `internal/ui/components.go`

**Purpose:** Define item types that wrap domain objects and implement `list.Item`.

```go
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
```

---

## Phase 2: Core TUI

### Task 2.1: Create `internal/ui/tui.go`

**Purpose:** Main Bubble Tea model with all interaction logic.

**Key Design Decisions:**
1. Use value receivers on `Init()` and `Update()` (Bubble Tea convention)
2. Store delegates in model, update `Focused` field only when focus changes
3. Track filter state per-panel for restoration on panel switch
4. Return `SelectionResult` via model field after quit

```go
package ui

import (
    "atelier-go/internal/config"
    "atelier-go/internal/locations"

    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

// Focus indicates which panel is active
type Focus int

const (
    FocusLocations Focus = iota
    FocusActions
)

// SelectionResult holds the final user selection
type SelectionResult struct {
    Location *locations.Location
    Action   *config.Action
    Canceled bool
}

// Model is the Bubble Tea model for the TUI
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
    width           int
    height          int
    quitting        bool
    lastLeftFilter  string
    lastRightFilter string
    focusChanged    bool // Track if focus changed this update

    // Result
    Result SelectionResult
}

// NewModel creates a TUI model from locations
func NewModel(locs []locations.Location) Model {
    // Convert to list items
    items := make([]list.Item, len(locs))
    for i, loc := range locs {
        items[i] = LocationItem{Location: loc}
    }

    // Delegates
    locDelegate := NewLocationDelegate()
    actDelegate := NewActionDelegate()

    // Location list
    locList := list.New(items, locDelegate, LeftWidth, ListHeight)
    locList.SetShowTitle(false)
    locList.SetShowFilter(false)
    locList.SetShowStatusBar(false)
    locList.SetShowHelp(false)

    // Action list (empty initially)
    actList := list.New(nil, actDelegate, RightWidth, ListHeight)
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
    ti.Width = LeftWidth
    ti.PromptStyle = lipgloss.NewStyle().Foreground(ColorBorder).Bold(true)
    ti.TextStyle = lipgloss.NewStyle().Bold(true)
    ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(ColorDimmed)

    return Model{
        allLocations:      locs,
        locations:         locList,
        locationsDelegate: locDelegate,
        actions:           actList,
        actionsDelegate:   actDelegate,
        filterInput:       ti,
        focus:             FocusLocations,
    }
}

func (m Model) Init() tea.Cmd {
    return tea.Batch(textinput.Blink, m.updateActions())
}

func (m Model) updateActions() tea.Cmd {
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
        items = []list.Item{ActionItem{
            Action: config.Action{Name: "No actions"},
        }}
    }

    return m.actions.SetItems(items)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    m.focusChanged = false

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c":
            m.Result = SelectionResult{Canceled: true}
            m.quitting = true
            return m, tea.Quit

        case "esc":
            return m.handleEscape()

        case "enter", "tab":
            return m.handleSelect()

        case "alt+enter", "ctrl+s":
            return m.handleFastSelect()

        case "up", "ctrl+p":
            return m.handleCursorUp()

        case "down", "ctrl+n":
            return m.handleCursorDown()
        }

    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        InitStyles(m.width, m.height)
        m.updateDimensions()
    }

    // Update filter input
    var cmd tea.Cmd
    m.filterInput, cmd = m.filterInput.Update(msg)
    cmds = append(cmds, cmd)

    // Sync filter to active list
    cmds = append(cmds, m.syncFilter()...)

    // Route non-key messages to both lists
    if _, isKey := msg.(tea.KeyMsg); !isKey {
        m.locations, cmd = m.locations.Update(msg)
        cmds = append(cmds, cmd)
        m.actions, cmd = m.actions.Update(msg)
        cmds = append(cmds, cmd)
    }

    return m, tea.Batch(cmds...)
}

func (m Model) handleEscape() (tea.Model, tea.Cmd) {
    if m.focus == FocusActions {
        m.focus = FocusLocations
        m.focusChanged = true
        m.filterInput.SetValue(m.lastLeftFilter)
        m.filterInput.Prompt = IconSearch + " "
        m.actions.SetFilterText("")
        m.lastRightFilter = ""
        return m, m.updateActions()
    }

    if m.filterInput.Value() == "" {
        m.Result = SelectionResult{Canceled: true}
        m.quitting = true
        return m, tea.Quit
    }

    // Clear filter
    m.filterInput.SetValue("")
    m.locations.SetFilterText("")
    m.lastLeftFilter = ""
    return m, m.updateActions()
}

func (m Model) handleSelect() (tea.Model, tea.Cmd) {
    if m.focus == FocusLocations {
        sel, ok := m.locations.SelectedItem().(LocationItem)
        if !ok {
            return m, nil
        }

        if sel.HasActions() {
            // Drill into actions
            m.focus = FocusActions
            m.focusChanged = true
            m.lastLeftFilter = m.filterInput.Value()
            m.filterInput.SetValue("")
            m.filterInput.Prompt = "Action " + IconSearch + " "
            m.actions.Select(0)
            return m, nil
        }

        // No actions - select with default
        loc := sel.Location
        m.Result = SelectionResult{Location: &loc, Action: nil}
        m.quitting = true
        return m, tea.Quit
    }

    // In actions panel
    locItem, _ := m.locations.SelectedItem().(LocationItem)
    actItem, ok := m.actions.SelectedItem().(ActionItem)
    if !ok {
        return m, nil
    }

    loc := locItem.Location
    act := actItem.Action
    m.Result = SelectionResult{Location: &loc, Action: &act}
    m.quitting = true
    return m, tea.Quit
}

func (m Model) handleFastSelect() (tea.Model, tea.Cmd) {
    sel, ok := m.locations.SelectedItem().(LocationItem)
    if !ok {
        return m, nil
    }

    loc := sel.Location
    m.Result = SelectionResult{Location: &loc, Action: nil}
    m.quitting = true
    return m, tea.Quit
}

func (m Model) handleCursorUp() (tea.Model, tea.Cmd) {
    if m.focus == FocusLocations {
        m.locations.CursorUp()
        return m, m.updateActions()
    }
    m.actions.CursorUp()
    return m, nil
}

func (m Model) handleCursorDown() (tea.Model, tea.Cmd) {
    if m.focus == FocusLocations {
        m.locations.CursorDown()
        return m, m.updateActions()
    }
    m.actions.CursorDown()
    return m, nil
}

func (m *Model) syncFilter() []tea.Cmd {
    var cmds []tea.Cmd
    val := m.filterInput.Value()

    if m.focus == FocusLocations {
        if val != m.lastLeftFilter {
            m.locations.SetFilterText(val)
            m.lastLeftFilter = val
            m.locations.Select(0)
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

func (m *Model) updateDimensions() {
    m.locations.SetSize(LeftWidth, ListHeight)
    m.actions.SetSize(RightWidth, ListHeight)
    m.filterInput.Width = LeftWidth
}

func (m Model) View() string {
    if m.quitting {
        return ""
    }

    // Update delegates only when focus changed (optimization)
    m.locationsDelegate.Focused = (m.focus == FocusLocations)
    m.locations.SetDelegate(m.locationsDelegate)
    m.actionsDelegate.Focused = (m.focus == FocusActions)
    m.actions.SetDelegate(m.actionsDelegate)

    search := SearchInputStyle.Render(m.filterInput.View())

    // Spotlight: only show panels when there's input or in actions mode
    if m.filterInput.Value() == "" && m.focus == FocusLocations {
        content := WindowStyle.Render(search)
        if m.width == 0 {
            return content
        }
        return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
    }

    // Full panel view
    left := LeftPanelStyle.Render(m.locations.View())

    actionTitle := "AVAILABLE ACTIONS"
    if m.focus == FocusActions {
        actionTitle = FocusedTitleStyle.Render("SELECT ACTION")
    }

    right := RightPanelStyle.Render(
        lipgloss.JoinVertical(lipgloss.Left, actionTitle, "", m.actions.View()),
    )

    inner := lipgloss.JoinVertical(
        lipgloss.Left,
        search,
        lipgloss.JoinHorizontal(lipgloss.Top, left, right),
    )

    content := WindowStyle.Render(inner)

    if m.width == 0 {
        return content
    }

    return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
```

---

## Phase 3: Integration

### Task 3.1: Update `internal/ui/ui.go`

Replace `runSelection` to use the new TUI instead of fzf:

```go
// runSelection executes the TUI and returns a session target
func runSelection(locs []locations.Location, cfg *config.Config) (*sessions.Target, error) {
    model := NewModel(locs)

    p := tea.NewProgram(model, tea.WithAltScreen())
    finalModel, err := p.Run()
    if err != nil {
        return nil, fmt.Errorf("TUI error: %w", err)
    }

    m, ok := finalModel.(Model)
    if !ok {
        return nil, fmt.Errorf("unexpected model type")
    }

    if m.Result.Canceled || m.Result.Location == nil {
        return nil, nil // User cancelled
    }

    // Resolve selection to session target
    shell := env.DetectShell()
    sessionManager := sessions.NewManager()

    actionName := ""
    if m.Result.Action != nil {
        actionName = m.Result.Action.Name
    }

    return sessionManager.Resolve(*m.Result.Location, actionName, shell, cfg.GetEditor())
}
```

### Task 3.2: Remove `showActionMenu` function

The action menu is now handled within the TUI's dual-panel interface.

### Task 3.3: Delete `internal/ui/fzf.go`

Remove the fzf dependency entirely.

### Task 3.4: Update imports in `ui.go`

Add Bubble Tea import, remove fzf-related code.

---

## Phase 4: Polish

### Task 4.1: Empty State Handling

In `tui.go`, handle edge cases:
- No locations: Show message "No projects found"
- No actions: Already handled with "No actions" placeholder

### Task 4.2: Help Footer (Optional)

Add a help line at the bottom of the window:
```
Enter:Select  Tab:Actions  Alt+Enter:Fast  Esc:Back  Ctrl+C:Quit
```

### Task 4.3: Manual Testing Checklist

- [ ] Empty location list
- [ ] Single location (project)
- [ ] Single location (folder)
- [ ] Project with 0 actions
- [ ] Project with 1 action
- [ ] Project with 5+ actions
- [ ] 100+ locations (performance)
- [ ] Small terminal (< 80 columns)
- [ ] Large terminal (> 200 columns)
- [ ] Resize terminal while running
- [ ] Filter, select, back, filter again
- [ ] Fast select (Alt+Enter)

---

## Phase 5: Cleanup

### Task 5.1: Remove POC

Delete `cmd/poc-tui/main.go` or move to `examples/poc-tui/`.

### Task 5.2: Update README

If applicable, add screenshots of the new TUI.

### Task 5.3: Mark Complete

Update this plan's checklist to mark all items complete.

---

## Checklist

### Phase 1: Foundation
- [ ] Create `internal/ui/styles.go`
- [ ] Create `internal/ui/components.go`
- [ ] Verify builds with `go build ./...`

### Phase 2: Core TUI
- [ ] Create `internal/ui/tui.go`
- [ ] Test standalone with mock data
- [ ] Verify all keybindings work

### Phase 3: Integration
- [ ] Update `internal/ui/ui.go` to use TUI
- [ ] Remove `showActionMenu` function
- [ ] Delete `internal/ui/fzf.go`
- [ ] Test with real location data
- [ ] Verify session attachment works

### Phase 4: Polish
- [ ] Test empty states
- [ ] Test edge cases from checklist
- [ ] Add help footer (optional)

### Phase 5: Cleanup
- [ ] Remove/move POC file
- [ ] Update documentation
- [ ] Mark plan complete

---

## Reference

- **POC Implementation**: `cmd/poc-tui/main.go`
- **Existing UI**: `internal/ui/ui.go`, `internal/ui/fzf.go`
- **Domain Types**: `internal/locations/locations.go`, `internal/config/config.go`
