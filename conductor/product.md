# Product Definition: Atelier Go

## Initial Concept
Atelier Go is a workspace launcher and session manager that focuses on supporting a native terminal experience free of multiplexers like Tmux. It allows you to quickly jump into projects and apps, making it easier to work with native terminal windows, tabs, and splits.

## Target Users
- **Native Terminal Enthusiasts:** Developers who prefer the native UI and integration of terminal emulators like Ghostty, iTerm2, or GNOME Terminal over terminal multiplexers.
- **Remote Workers:** Users who frequently work on remote servers via SSH and need stable, persistent sessions that survive network interruptions.
- **Efficiency Seekers:** Users who utilize tools like `zoxide` and want a unified, fuzzy-searchable interface to jump between their most frequent and configured locations.

## Core Goals
- **Seamless Native Experience:** Provide a workspace management workflow that feels like a first-class citizen of the operating system, utilizing native terminal features instead of abstracting them away.
- **Unified Workflow:** Bridge the gap between local and remote development by offering a consistent interface and set of actions regardless of where the code lives.
- **Robust Session Management:** Offer persistent session management that integrates deeply with the terminal and OS, ensuring work is never lost and recovery is instantaneous.

## Key Features
- **Unified Discovery:** A high-performance interactive UI that aggregates configured projects and `zoxide` directories into a single, fuzzy-searchable list.
- **Persistent Sessions:** Leverages `zmx` for session management, providing persistent environments that can be attached to, detached from, or recovered automatically (especially useful for SSH connections).
- **Customizable Actions:** A flexible action system allowing users to define specific commands (e.g., "Build", "Test", "Run Server") globally or on a per-project basis.
- **Remote Recovery:** Built-in support for session recovery via unique client IDs, allowing `autossh` to seamlessly restore a specific workspace after a drop.

## Configuration & Extensibility
- **YAML-driven Config:** Simple and powerful configuration using YAML for defining projects, UI themes, and global settings.
- **Multi-Host Support:** Support for host-specific configuration files, allowing a single dotfile repository to manage different project sets across multiple machines.
- **Hybrid Interface:** A rich Bubble Tea-based TUI for interactive use, complemented by a robust CLI for scriptability and direct session management.

## Long-term Vision
Atelier Go aims to become the primary entry point for all terminal-based work, local and remote, by providing the fastest and most reliable way to enter a focused development context.
