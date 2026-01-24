# Technology Stack: Atelier Go

## Core Technologies
- **Go (Golang):** The primary programming language, chosen for its performance, static typing, and excellent support for building CLI tools and concurrent processes.
- **Bubble Tea Ecosystem:**
    - **Bubble Tea (`github.com/charmbracelet/bubbletea`):** The functional framework for building the interactive TUI.
    - **Bubbles (`github.com/charmbracelet/bubbles`):** Standard UI components (like the text input used for fuzzy searching).
    - **Lip Gloss (`github.com/charmbracelet/lipgloss`):** The CSS-like styling library for terminal UI components.
- **Cobra (`github.com/spf13/cobra`):** The standard CLI framework for Go, used to manage commands, subcommands, and flags.
- **Viper (`github.com/spf13/viper`):** A comprehensive configuration solution for Go, used to handle YAML-based configuration files and host-specific overrides.

## External Dependencies & Integrations
- **ZMX:** The external session manager used to create, attach, and manage persistent terminal sessions across Linux and macOS.
- **Zoxide:** Integrated to provide fast directory discovery and jump capabilities based on user frequency and recency.
- **Fuzzy Search:** Powered by `github.com/sahilm/fuzzy` for high-performance string matching within the UI.

## Architecture
- **TUI-CLI Hybrid:** A single binary that provides both a rich interactive interface and a command-driven CLI for scriptability.
- **Action-Based Launcher:** A core logic that resolves locations (projects/zoxide) and maps them to executable actions, leveraging persistent sessions where possible.
