# Session Switching

GitHub Issue: https://github.com/jeremytondo/atelier-go/issues/41

## Overview

Add the ability to switch to a different zmx session when already inside a running session. Currently, attempting to attach to a new session from within an existing session fails silently.

## Technical Findings

1. **ZMX_SESSION environment variable**: zmx sets `ZMX_SESSION=<session-name>` when inside a session - this can be used to detect if we're in a session
2. **zmx detach command**: Detaches the client from the current session
3. **No native session switching**: zmx doesn't support attaching to a different session while inside one
4. **New zmx run command** (v0.2.0): Can create/run commands in a session without attaching, but attachment is still the blocker

## Solution: Coordinator Loop

The solution is a "coordinator loop" that wraps session attachment. When a user selects a location from the UI:

- **If outside a session**: Enter a coordinator loop that attaches, and after each detach checks for a pending switch request
- **If inside a session**: Write a pending switch file and run `zmx detach`, which returns control to the outer coordinator

### Flow Diagram

```
atelier-go
    |
    +-> [Outside session] Show UI -> user selects -> coordinator loop
    |
    +-> [Inside session] Show UI -> user selects -> write pending -> detach
    
Coordinator Loop:
    for {
        attach(target)
        // blocks until detach
        
        pending := readPendingSwitch()
        if pending == nil {
            break  // normal detach, exit to shell
        }
        target = pending  // switch, continue loop
    }
```

### Behavior Summary

| Context | Behavior |
|---------|----------|
| Outside session, run `atelier-go` | Show UI -> enter coordinator loop -> attach |
| Inside session, run `atelier-go` | Show UI -> write pending -> detach -> outer coordinator picks up |
| Normal detach (Ctrl+B,D) | Exit coordinator, return to parent shell |
| Cancel UI (outside session) | Exit immediately, no attach |
| Cancel UI (inside session) | Exit inner instance, stay in current session |

### Pending Switch File

Location: `/tmp/atelier-go-switch-<username>`

Format (JSON):
```json
{"name":"project:action","path":"/path/to/project","command":["zsh","-l","-i","-c","npm start"]}
```

## Checklist

### Task 1: Add Pending Switch File Helpers to `sessions` Package

- [ ] Add `InSession() bool` - Returns true if `ZMX_SESSION` env var is set
- [ ] Add `Detach() error` - Runs `zmx detach`
- [ ] Add `WritePendingSwitch(target *Target) error` - Writes target info to `/tmp/atelier-go-switch-<username>` as JSON
- [ ] Add `ReadPendingSwitch() (*Target, error)` - Reads and returns pending target (nil if none or file doesn't exist)
- [ ] Add `ClearPendingSwitch() error` - Removes the pending switch file

### Task 2: Refactor `ui.Run()` to Separate Selection from Attachment

- [ ] Rename/refactor `ui.Run()` to return `*sessions.Target` (or nil if cancelled) instead of handling attachment
- [ ] Move the attachment logic out of `ui.Run()` into the CLI layer
- [ ] Ensure UI package only handles user interaction, not session management

### Task 3: Update CLI Root Command with Coordinator Logic

- [ ] Detect if running inside a session using `sessions.InSession()`
- [ ] If inside session: write pending switch and detach
- [ ] If outside session: run coordinator loop with attach/pending-check cycle
- [ ] Handle user cancellation appropriately in both contexts

### Task 4: Update `sessions attach` Subcommand

- [ ] Update `internal/cli/sessions.go` to use the same coordinator loop pattern for consistency

### Task 5: Update `ui` Subcommand

- [ ] Update `internal/cli/ui.go` to use the same coordinator loop pattern for consistency

## Implementation Notes

### Files to Modify

| File | Changes |
|------|---------|
| `internal/sessions/sessions.go` | Add `InSession()`, `Detach()`, `WritePendingSwitch()`, `ReadPendingSwitch()`, `ClearPendingSwitch()` |
| `internal/ui/ui.go` | Refactor `Run()` to separate selection from attachment, return `*sessions.Target` |
| `internal/cli/cli.go` | Add coordinator loop logic to root command |
| `internal/cli/sessions.go` | Update `attach` subcommand to use coordinator loop |
| `internal/cli/ui.go` | Update `ui` subcommand to use coordinator loop |

### Example Coordinator Loop Implementation

```go
func run(ctx context.Context, cfg *config.Config) error {
    mgr := setupLocationManager(cfg, false, false)
    sessionMgr := sessions.NewManager()
    
    // Inside a session? Handle switch request
    if sessions.InSession() {
        target, err := ui.SelectTarget(ctx, mgr, cfg)
        if err != nil {
            return err
        }
        if target == nil {
            return nil  // User cancelled, stay in current session
        }
        
        // Write pending switch and detach
        sessions.WritePendingSwitch(target)
        return sessionMgr.Detach()
    }
    
    // Outside session - run coordinator loop
    target, err := ui.SelectTarget(ctx, mgr, cfg)
    if err != nil {
        return err
    }
    if target == nil {
        return nil  // User cancelled
    }
    
    // Coordinator loop
    for {
        fmt.Printf("Attaching to session '%s' in %s\n", target.Name, target.Path)
        sessionMgr.Attach(target.Name, target.Path, target.Command...)
        
        // After detach, check for pending switch
        pending, _ := sessions.ReadPendingSwitch()
        if pending == nil {
            break  // Normal detach, exit to shell
        }
        sessions.ClearPendingSwitch()
        target = pending  // Switch to new target
    }
    
    return nil
}
```

## Out of Scope

- Handling users who attach via raw `zmx attach` instead of through atelier-go
- Multi-terminal edge cases (using single per-user pending file)
- Re-attaching to the same session (let existing behavior work as-is)
