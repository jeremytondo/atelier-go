# UI Refactor: Native TUI with Bubble Tea

## Goal
Replace the existing `fzf` dependency with a native, elegant Go TUI built using the Charm ecosystem. This will provide a more integrated experience, better performance, and a highly customizable "Spotlight-style" interface.

## Tech Stack
- **Bubble Tea**: The TUI framework (Model-Update-View).
- **Bubbles**: Standard components (list, textinput, spinner).
- **Lip Gloss**: Styling and layout definitions.

## Core UX (Spotlight Mode)
- **Minimal Start**: The application launches as a minimal, centered search box with a search icon (``).
- **Dynamic Expansion**: The UI dynamically expands vertically to show results as the user begins typing.
- **Master-Detail Layout**: 
    - **Left Panel (Master)**: Displays Locations (Projects/Folders).
    - **Right Panel (Detail)**: Displays context-aware Actions for the currently selected location (e.g., "Attach", "Open in Editor", "New Session").

## Navigation & Interaction
- **Universal Search**: Filtering is focused on the active panel.
- **Master to Detail**: Press `Enter` or `Tab` to move focus from the Locations list to the Actions list.
- **Fast Select**: Press `Alt+Enter` or `Ctrl+S` to execute the default action immediately without entering the detail view.
- **Back/Cancel**: `Esc` navigates back from detail to master, or cancels/exits if already at the master level.
- **List Navigation**: Supports `Ctrl+N`/`Ctrl+P` and Arrow Keys.

## Visual Style
- **Layout**: Centered placement on the screen with rounded borders.
- **Iconography**:
    - Project icons: ``
    - Folder icons: `` (uea83)
- **Theming**:
    - Pink/Purple selection highlights for the active element.
    - Dimmed/Greyed-out state for unfocused panels to maintain visual hierarchy.

---

## POC Review Summary

A proof-of-concept TUI was implemented at `cmd/poc-tui/main.go` to validate the design. The POC successfully demonstrates the core UX concepts and has been reviewed for production readiness.

### What Works Well
- Proper alternate screen buffer usage for clean terminal restoration
- Correct message routing to both list models
- Clean filter state tracking across panel transitions (saves/restores filter when switching)
- Good spotlight UX behavior (minimal start, expands on input)
- Dual-panel layout with visual focus indicators

### Issues Identified for Production

| Priority | Issue | Location | Description |
|----------|-------|----------|-------------|
| High | `SetDelegate()` every render | Line 271-276 | Expensive operation called on every View() - should cache delegate state |
| High | Hardcoded list dimensions | Lines 355, 361 | Breaks on small terminals - needs responsive calculation |
| Medium | Hardcoded widths | Lines 20, 25, 31 | Values (100, 60, 35) not responsive to terminal size |
| Medium | Pointer receivers on Init/Update | Lines 116, 134 | Unconventional for Bubble Tea - should use value receivers |
| Low | `style.Copy()` deprecated | Line 76 | Should use `style.Inherit()` instead |

### Architectural Concerns
- The `item` struct conflates domain data with view presentation
- No separation between `Location` items and `Action` items (both use same struct)
- Missing error handling and loading states for production use
- Filter sync logic is correct but could be cleaner with separate state tracking

---

## Implementation Strategy

### File Structure

```
internal/ui/
  ui.go          # Public API (Run function) - minimal changes
  tui.go         # NEW: Bubble Tea model, Update, View
  components.go  # NEW: Custom delegates and item types
  styles.go      # NEW: Lip Gloss style definitions
  fzf.go         # DEPRECATED: Keep for fallback or remove
  render.go      # Keep: Formatting utilities (FormatLocation, StripANSI)
```

### Architecture Overview
- **Abstraction**: Separate domain types (`LocationItem`, `ActionItem`) from the POC's generic `item` struct
- **Model**: Standard Bubble Tea pattern with value receivers on Init/Update
- **Delegates**: Cache delegate instances, only update focus state when it changes
- **Display**: Use Alternate Screen buffer, handle resize events for responsive layout
- **Responsiveness**: Calculate dimensions from `tea.WindowSizeMsg` rather than hardcoding

### Data Flow

```
ui.Run()
    |
    v
locations.Manager.GetAll() --> []locations.Location
    |
    v
convertToItems(locs) --> []LocationItem
    |
    v
NewTUIModel(items, config) --> TUIModel
    |
    v
tea.NewProgram(model).Run()
    |
    v
SelectionResult{Location, Action} --> sessions.Manager.Resolve()
```

---

## Checklist

### Phase 1: Foundation
- [x] Research Bubble Tea list filtering performance for large project counts
- [x] Validate POC demonstrates core UX concepts
- [ ] Create `internal/ui/styles.go` with responsive style definitions
- [ ] Create `internal/ui/components.go` with LocationItem, ActionItem, and delegates
- [ ] Add proper type conversions from `locations.Location`

### Phase 2: Core TUI
- [ ] Create `internal/ui/tui.go` with TUIModel (value receivers)
- [ ] Implement all key handlers (esc, enter, tab, arrows, fast select)
- [ ] Implement spotlight behavior (expand on input)
- [ ] Implement dual-panel focus management
- [ ] Handle terminal resize events with responsive dimensions
- [ ] Fix delegate caching (don't call SetDelegate every render)

### Phase 3: Integration
- [ ] Update `internal/ui/ui.go` to use TUIModel instead of fzf
- [ ] Update `SelectionResult` to work with `sessions.Manager.Resolve()`
- [ ] Test with real location data from providers
- [ ] Remove or deprecate `internal/ui/fzf.go`

### Phase 4: Polish
- [ ] Add empty state handling (no locations, no actions)
- [ ] Add help footer with keyboard shortcuts
- [ ] Test edge cases (no projects, single project, many projects)
- [ ] Test keyboard navigation thoroughly
- [ ] Test small terminals (< 80 columns) and large terminals (> 200 columns)
- [ ] Implement Nerd Font fallback detection

### Phase 5: Cleanup
- [ ] Remove POC file `cmd/poc-tui/main.go` (or move to `examples/`)
- [ ] Update README with new TUI screenshots if applicable
- [ ] Mark this plan as complete

---

## Implementation Notes

### From Original Plan
- The TUI should gracefully handle cases where Nerd Fonts are not available (fallback to simple characters).
- Ensure that the search input is always focused when the user starts typing, regardless of which panel is selected (unless explicitly navigating the list).
- Use `bubbles/list` for the panels with custom delegates for the Master-Detail look.

### From POC Review
- **Delegate Management**: Store delegate instances in the model struct, update their `Focused` field directly, and only call `SetDelegate()` when focus actually changes (not every render).
- **Responsive Dimensions**: Use `InitStyles(width, height)` pattern to recalculate all widths/heights on `tea.WindowSizeMsg`. Set minimum viable dimensions (e.g., 60x10) for very small terminals.
- **Type Separation**: `LocationItem` should wrap `locations.Location` and implement `list.Item`. `ActionItem` should wrap `config.Action`. This keeps domain logic separate from view logic.
- **Filter State**: Track `lastLeftFilter` and `lastRightFilter` separately to enable clean panel switching with filter restoration.
- **Style Deprecation**: Use `lipgloss.NewStyle()` and method chaining instead of `style.Copy()`.

### Testing Strategy
1. **Unit Tests**: Test `LocationItem` and `ActionItem` filter/title methods
2. **Integration Tests**: Mock location data, verify selection result structure
3. **Manual Testing**:
   - Empty location list
   - Single location
   - 100+ locations (performance)
   - Projects with 0, 1, 5+ actions
   - Folders (no actions)
   - Small terminal (< 80 columns)
   - Large terminal (> 200 columns)

---

## References
- POC Implementation: `cmd/poc-tui/main.go`
- Detailed Implementation Plan: `.plans/ui-refactor-implementation.md`
