# State Management & Auto-Recovery

## Context
The user uses `atelier-go` as a remote shell via `autossh`. When the connection drops and reconnects, `atelier-go` restarts at the main menu instead of reconnecting to the previous session.

## Goal
Implement an Auto-Recovery Mechanism to automatically re-attach to the last active session upon reconnection if the session was not cleanly exited.

## Proposed Solution

### 1. State Tracking
- Use a state file (e.g., `current_session`) in `XDG_STATE_HOME` (default `~/.local/state/atelier-go/`).
- Content: The ID of the currently attached session.

### 2. Logic Changes
- **Before Attach**: Write the target session ID to the state file.
- **After Attach (Clean Exit)**: Remove the state file. This happens when the user manually detaches or exits the shell.
- **Startup (Crash Recovery)**:
    - Check if the state file exists.
    - If it exists, read the session ID.
    - Verify if the session is still running in `zmx` (using `zmx list`).
    - If running: Automatically attach to it (skipping the UI).
    - If not running: Clean up the state file and proceed to normal UI.

## Checklist
- [ ] Implement `GetStateDir` helper in `internal/utils/utils.go`.
- [ ] Add `SaveState`, `ClearState`, `LoadState` methods in `internal/sessions/sessions.go`.
- [ ] Update `Attach` workflow in `internal/sessions/sessions.go` to save state before attaching and clear it after detaching.
- [ ] Update `internal/ui/ui.go` `Run` method to check for recovery state before showing the menu.
- [ ] Verify session existence using `zmx list` before auto-recovering.

## Implementation Notes
- **File Location**: Ensure we respect XDG Base Directory specification. Use `~/.local/state/atelier-go/current_session`.
- **Files to Modify**:
    - `internal/sessions/sessions.go`: State management logic.
    - `internal/ui/ui.go`: Startup logic.
    - `internal/utils/utils.go`: Path helpers.
- **Edge Cases**:
    - File permissions.
    - Corrupt state file (should fall back to menu).
    - Session ID in state file no longer exists in `zmx` (should clean up and fall back to menu).
