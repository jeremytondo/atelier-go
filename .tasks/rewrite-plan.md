# Atelier Go: Phase 1 (Local First) - Implementation Plan

## Phase 1: Local Workflow (The "Clean Slate" Rewrite)

### Step 0: The "Clean Slate" Archive
**Goal**: clear the working directory for the new architecture while preserving history.
- [x] **Action 0.1**: Create `legacy/` directory.
- [x] **Action 0.2**: Move existing `cmd/`, `internal/`, and `main.go` into `legacy/`.
- [x] **Action 0.3**: Run `go mod tidy` to clean up dependencies.

### Step 1: Configuration & Domain
**Goal**: Define the data structures and load the "Source of Truth" (Config).
- [x] **Action 1.1**: Create `internal/config/config.go`.
    - Define `Project` struct (`Name`, `Path`).
    - Implement `Load(path string) (*Config, error)` using `viper` (or stdlib if preferred) to parse `~/.config/atelier/config.toml`.
    - Handle missing config gracefully (return empty config).

### Step 2: System Adapters (The "Hands")
**Goal**: Implement wrappers for external tools (`zoxide`, `zmx`).
- [ ] **Action 2.1**: Create `internal/zoxide/zoxide.go`.
    - Implement `Query() ([]string, error)` that runs `zoxide query -l`.
    - Parse output into a string slice.
- [ ] **Action 2.2**: Create `internal/zmx/session.go`.
    - Implement `Attach(name string, dir string) error`.
    - Use `exec.Command("zmx", "attach", "-c", name)`.
    - **Crucial**: Connect `os.Stdin`, `os.Stdout`, `os.Stderr`.
    - Set `cmd.Dir = dir` (ensures new sessions start in the right place).

### Step 3: The Data Engine (The "Brain")
**Goal**: Aggregation logic that merges config and zoxide data.
- [ ] **Action 3.1**: Create `internal/engine/engine.go`.
    - Define `Item` struct (`Name`, `Path`, `Source` [Project|Zoxide]).
    - Implement `Fetch() ([]Item, error)`.
    - Logic: Load Config + Query Zoxide -> Merge -> Deduplicate (Config wins) -> Return.

### Step 4: UI & CLI (The Interface)
**Goal**: The user-facing command-line interface.
- [ ] **Action 4.1**: Create `internal/ui/fzf.go`.
    - Implement `Select(items []string) (string, error)`.
    - Use `exec.Command("fzf", ...)` with piped input/output.
    - Keep it simple for Phase 1 (no complex bindings yet).
- [ ] **Action 4.2**: Create `cmd/atelier/main.go` & `cmd/atelier/root.go`.
    - Initialize the `rootCmd` using Cobra.
    - Wire up the "Run" logic:
        1. `engine.Fetch()`
        2. Format for `ui.Select()` (e.g., `[Project] Name  /path/to/dir`)
        3. Parse selection.
        4. `zmx.Attach()`.
- [ ] **Action 4.3**: Create `cmd/atelier/debug.go`.
    - Add a hidden `debug` command to print the raw list from `engine.Fetch()` without `fzf`, useful for verification.
