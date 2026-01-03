# UI Refactor: Native TUI with Bubble Tea

## Overview

Replace the `fzf` dependency with a native Go TUI using the Charm ecosystem. The POC at `cmd/poc-tui/main.go` validated the UX. The production implementation is complete.

**Tech Stack:**
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - Components (list, textinput)
- `github.com/charmbracelet/lipgloss` - Styling

---

## File Structure

```
internal/ui/
  ui.go          # Public API - calls TUI
  tui.go         # Bubble Tea model, Update, View
  components.go  # LocationItem, ActionItem, delegates
  styles.go      # Lip Gloss style definitions
```

---

## Checklist

### Phase 1: Foundation
- [x] Create `internal/ui/styles.go`
- [x] Create `internal/ui/components.go`
- [x] Verify builds with `go build ./...`

### Phase 2: Core TUI
- [x] Create `internal/ui/tui.go`
- [x] Test standalone with mock data
- [x] Verify all keybindings work

### Phase 3: Integration
- [x] Update `internal/ui/ui.go` to use TUI
- [x] Remove `showActionMenu` function
- [x] Delete `internal/ui/fzf.go`
- [x] Test with real location data
- [x] Verify session attachment works

### Phase 4: Polish
- [x] Add empty state handling (no locations, no actions)
- [x] Add help footer with keyboard shortcuts
- [x] Test edge cases from checklist
- [x] Test keyboard navigation thoroughly
- [x] Test small terminals (< 80 columns) and large terminals (> 200 columns)
- [x] Implement Nerd Font fallback detection

### Phase 5: Cleanup
- [x] Remove/move POC file
- [x] Update documentation
- [x] Mark plan complete

---

## Reference

- **Domain Types**: `internal/locations/locations.go`, `internal/config/config.go`
