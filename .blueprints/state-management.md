# State Management & Auto-Recovery

## Context
The user uses `atelier-go` as a remote shell via `autossh`. When the connection drops and reconnects, `autossh` restarts the application. Without state management, the user is dropped back into the main menu instead of their active session.

## Goal
Implement a robust Auto-Recovery Mechanism to automatically re-attach to the last active session upon reconnection, supporting multiple concurrent terminals.

## Implementation Details

### 1. Client Identification
Since multiple terminal sessions may exist simultaneously, we use a `--client-id` flag to distinguish between them. This flag is typically provided by a local shell wrapper.

**Local Shell Wrapper (Mac):**
```bash
agw() {
    export ATELIER_CLIENT_ID="${ATELIER_CLIENT_ID:-$(uuidgen | cut -d'-' -f1)}"
    autossh -M 0 -q -t ag -- "atelier-go --client-id=$ATELIER_CLIENT_ID"
}
```

### 2. State Tracking
- **Location:** `~/.local/state/atelier-go/sessions/<client-id>`
- **Content:** The ID of the currently attached `zmx` session.

### 3. Logic Flow

- **Before Attach**: Write the target session ID to the state file.
- **After Attach (Clean Exit)**: Remove the state file.
- **Startup (Recovery Check)**:
    1. Check if `--client-id` flag is provided.
    2. If yes, check for state file at `~/.local/state/atelier-go/sessions/<client-id>`.
    3. If state file exists, read the session ID.
    4. Verify if the session is still running (using `zmx list`).
    5. If running: Automatically attach (skipping UI).
    6. If not running: Clean up state file and proceed to normal UI.

## Components Modified

- `internal/utils/utils.go`: Added `GetStateDir()` helper (XDG compliant).
- `internal/sessions/sessions.go`: Added `SaveState`, `LoadState`, `ClearState`, and `SessionExists`.
- `internal/ui/ui.go`: Updated `Run` to accept `clientID` and handle state lifecycle.
- `internal/cli/cli.go`: Added persistent `--client-id` flag and `tryRecover` logic.
- `internal/cli/ui.go`: Integrated recovery check into `ui` command.

## Edge Cases Handled
- **Connection Drop**: State file remains, triggering recovery on next run.
- **Session Termination**: If the remote session ends while disconnected, the state is cleaned up and user sees the menu.
- **Multiple Terminals**: Unique client IDs prevent state collisions.
- **Manual Attach**: Recovery only triggers when launching the UI/Root command, not when using explicit `sessions attach` commands.
