# Agents & Components

This document describes the architecture and components of the `atelier-go` system.

## Core System Architecture

The system is built as a unified Go binary (`atelier-go`) that operates as a local-first interactive CLI. It replaces the previous Bash-based implementation and the legacy Client/Server model.

### The CLI Agent (`atelier-go`)

The main entry point for the user.

- **Responsibilities:**
  - Aggregates projects from static configuration and dynamic `zoxide` history.
  - Provides a two-step interactive workflow using `fzf`:
    1.  **Project Selection**: Filter and select a target workspace.
    2.  **Action Selection**: (Optional) Choose a specific task to run (e.g., "Run Server", "Test").
  - Manages session lifecycle via `zmx` (a wrapper around `shpool`).

## Workflow

1.  **Launch**: User runs `atelier-go`.
2.  **Select**: The agent displays a fuzzy-searchable list of projects.
3.  **Act**:
    - If specific **Actions** are defined for the project, the user is prompted to select one (or default to "Shell").
    - If no actions are defined, it defaults to a standard shell session.
4.  **Attach**: The agent connects to a persistent session in the target directory, executing the selected command if provided.

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

The system orchestrates several external tools to provide robust functionality:

- **shpool:** Handles persistent shell sessions, ensuring work is not lost if the connection drops.
- **zmx:** An internal package that acts as the bridge to `shpool`.
- **zoxide:** Provides "frecent" (frequent + recent) directory tracking.
- **fzf:** Powers the interactive selection interface.

## Coding Standards

All code MUST adhere to standard Go conventions and pass linting checks.

1.  **Error Strings**: Error messages used with `fmt.Errorf` or `errors.New` must be lowercase and not end with punctuation (e.g., `fmt.Errorf("error reading file: %w", err)`).
2.  **Comments**: All exported functions, types, and variables must have documentation comments starting with the name of the exported entity.
3.  **Linting**: Run `go vet ./cmd/... ./internal/...` to verify correctness.
4.  **Formatting**: All code must be formatted with `gofmt`.

## Agent Behavior

- NEVER commit anything to git.
- DO NOT write tests.
