# Agents & Components

This document describes the architecture and components of the `atelier-go` app.

## RULES
* NEVER commit anything to git.
* DO NOT write tests.

## Core System Architecture

The system is built as a local-first interactive CLI, organized in a domain-driven flat structure.

### The CLI Agent (`atelier-go`)

The main entry point for the user.

- **Responsibilities:**
  - Aggregates locations via the **Location Provider** architecture (Projects, Zoxide).
  - Provides a two-step interactive workflow using `fzf`:
    1.  **Project Selection**: Filter and select a target workspace.
    2.  **Action Selection**: (Optional) Choose a specific task to run.
  - Manages session lifecycle via `sessions` (wrapping `zmx`/`shpool`).

### Server Agent (`atelier-go-server`)
- Stubbed for future remote agent capabilities.

## Codebase Structure

The project follows a standard Go CLI layout with a flat internal structure:

- **`cmd/`**: Binary entry points (`atelier-go`, `atelier-go-server`).
- **`internal/`**:
  - **`cli/`**: Cobra command definitions.
  - **`config/`**: Configuration loading and structs.
  - **`locations/`**: Location discovery via Providers (`Provider` interface, `Manager`).
  - **`sessions/`**: Session management (`Manager`, `Session`).
  - **`ui/`**: Interactive UI logic (`fzf` wrapper, rendering).
  - **`utils/`**: Shared helpers.

## Workflow

1.  **Launch**: User runs `atelier-go`.
2.  **Discovery**: The `locations` manager queries all providers (Config, Zoxide) in parallel.
3.  **Select**: The `ui` package displays a merged, fuzzy-searchable list.
4.  **Act**: User selects a location and optionally an action.
5.  **Attach**: The `sessions` manager attaches to the persistent session.

## Configuration

Configuration is stored in `~/.config/atelier-go/config.toml`.

```toml
# Example Configuration

[[projects]]
  name = "atelier-go"
  path = "/Users/me/Projects/atelier-go"

  [[projects.actions]]
    name = "Run Server"
    command = "go run main.go server"

  [[projects.actions]]
    name = "Test"
    command = "go test ./..."
```

## External Helper Agents

The system orchestrates several external tools:

- **zmx:** Session manager wrapping `shpool`.
- **zoxide:** Directory jumper (used as a Location Provider).
- **fzf:** Fuzzy finder for the UI.

## Coding Standards

All code MUST adhere to standard Go conventions and pass linting checks.

1.  **Package Entry Points**: Each package must have a primary file named after the package (e.g., `locations/locations.go`) containing the core types and constructors.
2.  **Error Strings**: Error messages used with `fmt.Errorf` or `errors.New` must be lowercase and not end with punctuation.
3.  **Comments**: All exported entities must be documented. Packages must have package comments.
4.  **Linting**: Run `go vet ./cmd/... ./internal/...` to verify correctness.
5.  **Formatting**: All code must be formatted with `gofmt`.
