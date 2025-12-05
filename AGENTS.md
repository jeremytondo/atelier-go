# Agent Guidelines for Atelier Go

## Build & Test
- **Linting**: Use `shellcheck` for static analysis.
  ```bash
  shellcheck atelier-client atelier-server
  ```
- **Testing**: Currently manual. See `README.md` for local testing instructions.
  - Ensure `ssh`, `fzf` (client), and `shpool`, `zoxide` (server) are installed.
- **Execution**: Run scripts directly from source or symlink to path.

## Code Style & Conventions
- **Language**: Bash (use `#!/usr/bin/env bash`).
- **Safety**: ALWAYS start scripts with `set -euo pipefail`.
- **Formatting**: 
  - Indentation: **2 spaces**.
  - Use `[[ ... ]]` for conditions, not `[ ... ]`.
- **Naming**:
  - Functions: `snake_case` (e.g., `cmd_interactive_launch`, `check_dependencies`).
  - Globals/Config: `UPPER_CASE` (e.g., `HOST`, `SERVER_PATH`).
  - Locals: `lower_case` declared with `local` (e.g., `local location`).
- **Structure**:
  - Group functions by category (Config, Utilities, Commands).
  - Use a `main` function at the end to parse arguments.
  - Use `usage` function for help text.
- **Comments**:
  - **File Headers**: Use `=` separators for script descriptions.
  - **Sections**: Use `--- Section Name ---` to divide code sections.
  - **Logic**: Use numbered steps (e.g., `# 1. Step`) for complex logic.
- **Error Handling**:
  - Exit with non-zero status on fatal errors.
