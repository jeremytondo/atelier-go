# Action Trigger & Editor Launch

## Overview
Implement a dual-trigger system for the main location list to support efficient workflows for both simple folders and configured projects.

## Interaction Model

We define two triggers for the main selection list:
1.  **Primary Trigger (`Enter`)**: The default "do what I mean" action.
2.  **Secondary Trigger (`Alt-Enter`)**: The alternative or "explore" action.

### Behavior Matrix

| Location Type | Primary (`Enter`) | Secondary (`Alt-Enter`) |
| :--- | :--- | :--- |
| **Folder** (Zoxide) | Open Shell | Open Editor |
| **Project** (Configured) | Run Default Action* | Open Action Menu |

* *Default Action* is the first action defined in the project config. If no actions are defined, fallback to opening a shell.

## Implementation Details

### 1. Configuration (`internal/config`)
Add a new field to the `Config` struct to specify the preferred editor.

```go
type Config struct {
    // ... existing fields
    Editor string `mapstructure:"editor"`
}
```

**Resolution Logic:**
When "Open Editor" is triggered, resolve the editor command in this order:
1.  `config.yaml` value (`editor`).
2.  `$EDITOR` environment variable.
3.  Fallback: `vim`.

*Note: For this iteration, the editor setting does not support arguments (e.g., `code -r`). It is treated as a command name.*

### 2. UI Updates (`internal/ui`)
- **FZF Integration:** Update `Select` to listen for `alt-enter`.
- **Selection Logic:** Refactor `runSelection` to handle the `key` returned by FZF.
    - Remove the existing `alt-s` specific logic.
    - Implement the Behavior Matrix logic to determine the `sessionName` and `commandArgs`.

### 3. Execution
- **Open Editor:** Create a session where the command is `<editor_cmd> .` (opening the current directory).
- **Default Action:** Auto-select `actions[0]` without showing the secondary menu.
- **Action Menu:** Reuse the existing secondary menu logic to let the user choose a specific action.

## Checklist
- [x] Update `Config` struct in `internal/config/config.go`.
- [x] Implement editor resolution logic.
- [x] Update `fzf.go` to bind and expect `alt-enter`.
- [x] Refactor `runSelection` in `ui.go` to implement the trigger matrix.
- [x] Remove legacy `alt-s` handling.
- [x] Verify behavior for Folders (Shell vs Editor).
- [x] Verify behavior for Projects (Default Action vs Menu).

## Implementation Notes

- **Configuration:** The `Config` struct in `internal/config/config.go` was updated to include an `Editor` field, mapped from the YAML config.
- **CLI Setup:** `internal/cli/setup.go` was refactored to accept an existing `*config.Config` object. This avoids redundant file loading and ensures consistency across the application.
- **Trigger Matrix:** The logic in `internal/ui/ui.go` was updated to implement the behavior matrix using `Alt-Enter` as the secondary key.
- **Cleanup:** All legacy `Alt-s` support and related logic were removed from the codebase.
