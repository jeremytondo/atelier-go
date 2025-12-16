# Atelier Go

Atelier Go is a CLI tool that streamlines development on remote machines by bridging the gap between your local terminal's native UI and your remote workflows. It acts as a session manager and launcher, allowing you to quickly jump into `shpool` sessions or projects on a remote server directly from your local terminal.

## Table of Contents

- [About](#about)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
  - [File Locations](#file-locations)
  - [Local Overrides](#local-overrides)
  - [Client Options](#client-options)
  - [Server Options](#server-options)
  - [Projects](#projects)
- [Usage](#usage)
  - [Client](#client)
  - [Server](#server)
- [Installation Details](#installation-details)

## About

I was inspired by the [Ghostty](https://ghostty.org/) terminal's dedication to native UI. The way they integrate with the OS—windows, tabs, splits—just feels right. However, this "native" feeling often breaks down when working on remote machines. You usually end up relying on multiplexers like Tmux or Zellij to manage sessions, which creates a layer of separation from your terminal's native features.

**Atelier Go** solves this by acting as a bridge. It runs a daemon on your remote machine and a client on your local machine. The client presents an interactive, fuzzy-searchable list of your remote sessions, projects, and frequent directories. Selecting one instantly launches or attaches to a session in your current terminal window. This allows you to use your terminal's native tabs and windows to manage remote workspaces instead of a nested multiplexer.

## Quick Start

1.  **Download:** Grab the latest release for your OS from the [GitHub Releases page](https://github.com/jeremytondo/atelier-go/releases).
2.  **Install:** Ensure the binary is in your `PATH` on both your **local** (client) and **remote** (server) machines.
3.  **Start Server (Remote):**
    ```bash
    atelier-go server start
    ```
    *Note the token displayed in the output.*

4.  **Connect (Local):**
    ```bash
    atelier-go client login <token>
    ```

5.  **Launch:**
    ```bash
    atelier-go client
    ```

## Configuration

Atelier Go uses TOML files for configuration.

### File Locations

Configuration files are stored in `~/.config/atelier-go/` (or `$XDG_CONFIG_HOME/atelier-go/`).

- **Client:** `~/.config/atelier-go/client.toml`
- **Server:** `~/.config/atelier-go/server.toml`

### Local Overrides

You can create "local" configuration files to override settings without changing the main config file. This is useful for machine-specific settings (like a different host or port) that you don't want to sync across dotfiles.

- **Client Override:** `~/.config/atelier-go/client.local.toml`
- **Server Override:** `~/.config/atelier-go/server.local.toml`

Values in the `.local.toml` file will take precedence over the base `.toml` file.

### Client Options

**`~/.config/atelier-go/client.toml`**

```toml
# The remote host to connect to (default: localhost)
host = "my-dev-server.com"

# The port the server is listening on (default: 9001)
port = 9001

# Default filter view: "sessions", "projects", "frequent", or "all"
default-filter = "projects"

# Custom keybindings for the interactive selector
[keys]
sessions = "ctrl-s"
projects = "ctrl-p"
frequent = "ctrl-f"
all = "ctrl-a"
```

### Server Options

**`~/.config/atelier-go/server.toml`**

```toml
# Address to bind to (default: 0.0.0.0)
host = "0.0.0.0"

# Port to listen on (default: 9001)
port = 9001

# Define custom actions available when creating a new session
[[actions]]
name = "Edit (Neovim)"
command = "nvim"

[[actions]]
name = "Shell"
command = "$SHELL -l"
```

### Projects

You can define projects by creating TOML files in the `~/.config/atelier-go/projects/` directory. Each file represents a separate project.

**Example: `~/.config/atelier-go/projects/my-app.toml`**

```toml
name = "My Application"
location = "~/dev/my-app"

[[actions]]
name = "Run Server"
command = "npm start"

[[actions]]
name = "Run Tests"
command = "npm test"
```

*   **`name`**: The display name of the project shown in the client.
*   **`location`**: The absolute path to the project directory (supports `~` expansion).
*   **`actions`**: A list of custom commands available for this project.

## Usage

### Client

The client is your main interface. Running `atelier-go client` opens an interactive `fzf` window.

**Interactive Controls:**
- **Enter:** Select item (attach to session or start new one).
- **Ctrl-S:** Switch to **Active Sessions**.
- **Ctrl-P:** Switch to **Projects**.
- **Ctrl-F:** Switch to **Frequent Directories** (zoxide).
- **Ctrl-A:** Switch to **All Directories**.

**CLI Flags:**
- `atelier-go client --sessions` (Start directly in Sessions view)
- `atelier-go client --projects` (Start directly in Projects view)
- `atelier-go client --all` (List all directories)

### Server

The server runs as a background daemon on your remote machine.

- **Start:** `atelier-go server start` (Runs in background)
- **Stop:** `atelier-go server stop`
- **Status:** `atelier-go server status`
- **Restart:** `atelier-go server restart`
- **Token:** `atelier-go server token` (Show current auth token)

## Installation Details

### Systemd Integration (Linux)

You can install the server as a systemd user service so it starts automatically on boot/login.

```bash
atelier-go server install
```
This will create and enable a service file in `~/.local/share/systemd/user/`.
