# Attach Command Plan

I'd like to be able to easily use the CLI instead of the UI to create or attach
to sessions. This would help if a user wanted to script the opening of session
instead of using the ui.

Right now in sessions.go we have an Attach command, but it acts more as a simple
wrapper around zmx. Instead, I think this needs to be a bit more robust and work
the same as if a user had selected a location and action from the UI.

I think we'll want to support the following types of commands:

Open the default action of a project:
```bash
atelier-go session attach --project <projectname>
```

Open a specific action of a project:
```bash
atelier-go session attach --project <projectname> --action <actionname>
```


Open a folder:
```bash
atelier-go session attach --folder <folderpath>
```

We should also support shorter flags like -p, -a, and -f.

These should work the same as if a user had selected them directly from the UI.
They should create a new session if one doesn't exist, or attach to an existing
session if it does exist.

## Implementation Plan

### 1. Refactor & Centralize Logic
Move the session resolution logic (converting a "Location + Action" into a "Session Config") from `ui` to the `sessions` package to ensure consistency between CLI and UI.

- **Create `sessions.Target` struct**: Replaces `WorkflowResult` in `ui.go`.
  ```go
  type Target struct {
      Name    string
      Path    string
      Command []string
  }
  ```
- **Create `sessions.Resolve(loc locations.Location, actionName string, shell string) (*Target, error)`**:
  - Encapsulate the "Logic Matrix" (Default actions, Shell wrapping).
  - Ensure consistent naming conventions (e.g., `project:action`).

### 2. Enhance Location Manager
Update `internal/locations/locations.go` to support direct lookups.

- **Add `Find(name string)` to `Manager`**:
  - Allows the CLI to find a project by name efficiently.

### 3. Implement CLI Flags
Update `internal/cli/sessions.go` to handle new flags for the `attach` command.

- **Flags**:
  - `--project`, `-p`: Name of the project.
  - `--action`, `-a`: (Optional) Specific action to run.
  - `--folder`, `-f`: Path to a generic folder.
- **Logic**:
  - If flags are present, resolve the location (via `Find` or path) and action (via `Resolve`).
  - **Decision**: Removed legacy positional argument support in favor of a cleaner flag-based CLI for the `attach` command.

### 4. Update UI
Refactor `internal/ui/ui.go` to use the new `sessions.Resolve` function, removing duplicated logic.

## Checklist

- [x] Refactor `ui` logic to `sessions` package
    - [x] Create `sessions.Target` struct
    - [x] Create `sessions.Resolve` function
- [x] Enhance `Location Manager`
    - [x] Add `Find(name string)` to `Manager`
- [x] Update `cli/sessions.go`
    - [x] Add flags: `--project`, `--action`, `--folder`
    - [x] Implement logic to use `Find` and `Resolve`
- [x] Update `ui/ui.go`
    - [x] Use `sessions.Resolve` instead of inline logic

## Implementation Notes

- **"Shell" Reserved Action**: "Shell" is now a reserved action name in `sessions.Resolve`. This allows explicitly launching a shell for projects that have defined actions, which would otherwise default to their first defined action.
- **Legacy Arguments**: Legacy positional argument support was removed to simplify the CLI interface and reduce ambiguity between project names and paths.
- **Architecture**: The logic was consolidated into `sessions.Manager.Resolve`. The `internal/cli/sessions.go` file was refactored into smaller, focused functions to improve readability and maintainability.
- Ensure that the session name generation is consistent with how the UI does it currently.
- The `Find` method in `Location Manager` should probably prioritize exact matches on project names.
