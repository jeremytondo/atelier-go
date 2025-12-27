# Atelier Go: Refactor & Location Provider Architecture

## 1. Context & Goal
We are reorganizing the atelier-go codebase to improve readability and follow Go best practices for CLI tools. We are moving away from abstract architectural patterns in favor of a straightforward, domain-driven flat structure using a Provider pattern for discovery.

## 2. New Directory Structure
Refactor the workspace to match this layout. Note: The legacy/ folder contains deprecated code and should be used only for reference; do not import from it.

```text
atelier-go/
├── cmd/
│   ├── atelier-go/           # Client Entry Point
│   └── atelier-go-server/    # Server Entry Point (Stub)
├── internal/
│   ├── cli/                  # Cobra Command definitions (root.go, locations.go, sessions.go)
│   ├── config/               # Config parsing and Project structs
│   ├── locations/            # Location discovery logic
│   │   ├── location.go       # Location struct definition
│   │   ├── provider.go       # Provider interface
│   │   ├── projects.go       # Project-based provider
│   │   ├── zoxide.go         # Zoxide-based provider
│   │   └── manager.go        # Provider orchestration & deduplication
│   ├── sessions/             # zmx persistence logic (session.go, zmx.go)
│   ├── ui/                   # interactive fzf wrapper (fzf.go, render.go, workflow.go)
│   └── utils/                # Shared helpers (paths.go)
└── legacy/                   # Archived reference code
```

## 3. Implementation Steps

### Step 1: Internal Logic & Providers
- `internal/utils/paths.go`: Implement `Shorten(path)` to handle tilde expansion (`~`).
- `internal/config/config.go`: Consolidate `Config` and `Project` structs.
- `internal/locations/`:
    - Define `Provider` interface in `provider.go`.
    - Implement `projects.go` as a `Provider`.
    - Implement `zoxide.go` as a `Provider` (encapsulating zoxide binary calls).
    - Implement `manager.go` to orchestrate providers and return merged, deduplicated locations.
- `internal/sessions/`:
    - Move all zmx and session logic here.
    - Define `Session` struct.
    - Implement `List`, `Attach`, and `Kill` functions.

### Step 2: UI & Workflow
- `internal/ui/fzf.go`: Low-level fzf execution wrapper.
- `internal/ui/render.go`: Formatting logic for Locations (icons, colors, shortened paths).
- `internal/ui/workflow.go`: Implement the high-level interactive flow:
    1. Fetch all locations.
    2. Select location via fzf.
    3. Select action (if multiple) or default to shell.
    4. Return target details.

### Step 3: CLI Refactor
- `internal/cli/`: Implement Cobra commands.
    - `root.go`: Default action launches `ui.InteractiveWorkflow()`.
    - `locations.go`: Command to print merged locations.
    - `sessions.go`: Subcommands for list, attach, and kill.

### Step 4: Binary Entry Points
- `cmd/atelier-go/main.go`: Initialize config and execute `cli.Execute()`.
- `cmd/atelier-go-server/main.go`: Basic stub entry point.

## 4. Review & Refactoring Instructions
1. **Readability**: Keep code simple and direct.
2. **Error Handling**: Use idiomatic Go error wrapping.
3. **Dependency Injection**: Pass Config/Dependencies explicitly.
4. **Binary Execution**: Robust `exec.Command` usage with proper Stdin/Out/Err handling.

## Checklist
- [ ] Create `internal/utils/paths.go` with `Shorten` function.
- [ ] Consolidate `internal/config/config.go`.
- [ ] Implement `internal/locations/` provider architecture.
- [ ] Move and refactor session logic to `internal/sessions/`.
- [ ] Implement UI components in `internal/ui/`.
- [ ] Set up Cobra commands in `internal/cli/`.
- [ ] Update `cmd/atelier-go/main.go` and create `cmd/atelier-go-server/main.go`.
- [ ] Verify all legacy code is moved or accounted for in the new structure.
- [ ] Ensure all code passes `go vet`.

## Implementation Notes
- The `Provider` interface should return a slice of `Location` objects and an error.
- Deduplication in `manager.go` should prefer `Config` projects over `zoxide` paths if they overlap.
- `ui.render.go` should handle terminal colors and icons to make the `fzf` list visually appealing.
- Ensure that the interactive workflow gracefully handles `fzf` being cancelled (non-zero exit code).
