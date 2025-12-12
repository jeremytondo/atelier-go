# Agents & Components

This document describes the architecture and components (Agents) of the `atelier-go` system, which replaces the previous Bash-based implementation with a unified Go application.

## Core System Agents

The system is built as a single Go binary (`atelier-go`) that operates in two primary modes:

### 1. The Daemon (Server)

**Command:** `atelier-go server`

The Daemon is the central agent running on the host machine (where your code lives).

- **Responsibilities:**
  - Runs as a background service (default port: `9001`).
  - Exposes a secure HTTP API protected by Bearer tokens.
  - Aggregates state from system integrations (`shpool` sessions, `zoxide` paths).
  - Handles health checks and authentication.

### 2. The Client

**Command:** `atelier-go client`

The Client is the user-facing agent that facilitates connection and workflow management.

- **Responsibilities:**
  - Connects to the Daemon via HTTP to fetch active sessions and recent locations.
  - Provides an interactive, fuzzy-searchable UI using `fzf`.
  - Manages the SSH connection lifecycle.
  - Attaches to existing `shpool` sessions or creates new ones based on user selection.

## External Helper Agents

The system orchestrates several external tools to provide robust functionality:

- **shpool:** Handles persistent shell sessions, ensuring work is not lost if the connection drops.
- **zoxide:** Provides "frecent" (frequent + recent) directory tracking, allowing quick navigation to commonly used project folders.
- **fzf:** Powers the interactive selection interface for filtering sessions and directories.
- **ssh:** Provides the secure transport layer for the client to interact with the host environment.

## Workspace Actions

When initiating a new session in a target directory, the user can choose from the following operational modes:

1. **Edit:** Launches `nvim` (Neovim) directly in the target directory.
2. **Shell:** Launches a standard interactive shell (`bash` or configured `$SHELL`).
3. **Opencode:** Launches the `opencode` CLI agent in the target directory.

## Configuration Conventions

- **Parameter Naming:** All keys in configuration files (e.g., `client.toml`) MUST use **kebab-case** (e.g., `default-filter`, `log-level`) instead of snake_case or camelCase.

## Agent Behavior

- NEVER commit anything to git.
