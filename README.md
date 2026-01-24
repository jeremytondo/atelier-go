# Atelier Go

Atelier Go is a workspace launcher and session manager that focuses on supporting a native terminal experience free of multiplexers like Tmux.
It allows you to quickly jump into projects and apps, making it easier to work with native terminal windows, tabs, and splits.

## Table of Contents

- [About](#about)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
  - [General Settings](#general-settings)
  - [Theme](#theme)
  - [Projects](#projects)
  - [Host-Specific Projects](#host-specific-projects)
- [Usage](#usage)
  - [Interactive UI](#interactive-ui)
  - [Sessions](#sessions)
  - [Locations](#locations)
- [Remote Work](#remote-work)

## About

I was inspired by the [Ghostty](https://ghostty.org/) terminal's dedication to native UI. The way it integrates with the OS—windows, tabs, splits—just feels right. Usually, this "native" feeling breaks down when you start juggling multiple projects or working on remote machines. You often end up relying on multiplexers like Tmux or Zellij, which adds a layer of complexity and separates you from your terminal's native features.

**Atelier Go** fixes this by making your local and remote workflows feel like first-class citizens. It aggregates your configured projects and `zoxide` directories into a single, fuzzy-searchable list. Selecting one instantly attaches to a persistent session or starts a new one. This lets you use your terminal's native tabs and windows to manage all your workspaces, no matter where they live.

## Quick Start

1.  **Install:** Download the appropriate binary and ensure `atelier-go` is in your `PATH`.
2.  **Dependencies:** You'll need `fzf`, `zmx`, and `zoxide` installed on your machine.
3.  **Launch:**
    ```bash
    atelier-go
    ```

## Configuration

Atelier Go looks for configuration in `~/.config/atelier-go/`.

### General Settings

You can define global settings like your preferred editor and default actions at the top of `config.yaml`.

```yaml
editor: "nvim"
shell-default: true

actions:
  - name: "Build"
    command: "make"
```

*   **`editor`**: The command used to open folders (e.g., `nvim`, `vim`, `code`). If not set, it defaults to the `$EDITOR` environment variable, then `vim`.
*   **`shell-default`**: If set to `true`, a "Shell" action is prepended to the beginning of the action list for all locations, making it the default. Defaults to `false` (Shell is appended to the end).
*   **`actions`**: A list of global actions that will be available for all discovered locations (projects and zoxide directories).

### Theme

You can customize the UI colors by adding a `theme` section to your `config.yaml`.

```yaml
theme:
  primary: "#89b4fa"         # Main window and search box borders
  accent: "#74c7ec"          # Project icons and names, search icon
  highlight: "#cba6f7"       # Selections and focused panel headers
  text: "#ffffff"            # Primary content and labels
  subtext: "240"             # Secondary info and help text
```

*   Colors can be specified as Hex strings (e.g., `"#RRGGBB"`) or ANSI color numbers (e.g., `"240"`).
*   **`primary`**: Used for all major UI borders and the search prompt icon.
*   **`accent`**: Used for project identifiers (icons/names).
*   **`highlight`**: Used for the currently selected item and focused panel headers.
*   **`text`**: Used for primary labels and standard text.
*   **`subtext`**: Used for help text, inactive borders, and secondary information.

### Projects

You can define projects by creating a `config.yaml` file in `~/.config/atelier-go/`.

**Example: `~/.config/atelier-go/config.yaml`**

```yaml
projects:
  - name: "My Application"
    path: "~/dev/my-app"
    default-actions: true
    shell-default: false
    actions:
      - name: "Run Server"
        command: "npm start"
```

*   **`name`**: The display name shown in the UI.
*   **`path`**: The directory to jump into (supports `~` expansion).
*   **`default-actions`**: Whether to include global actions for this project. Defaults to `true`.
*   **`shell-default`**: Override the global `shell-default` setting for this specific project.
*   **`actions`**: Custom commands for this project. These are merged with global actions if `default-actions` is `true`, with project-specific actions taking precedence.

### Host-Specific Projects

If you work across multiple machines, you can define projects that only show up on a specific host. Create a YAML file named after the host in the same directory:

`~/.config/atelier-go/<hostname>.yaml`

Settings in the host-specific file will be merged with the global `config.yaml`.

To see what hostname Atelier Go is using for your current host, run:

```bash
atelier-go hostname
```

You can override the detected hostname by setting the `ATELIER_HOSTNAME` environment variable.

## Environment Variables

Atelier Go supports several environment variables to customize its behavior:

| Variable | Description |
| :--- | :--- |
| **`EDITOR`** | The command used to open folders (e.g., `nvim`, `code`). |
| **`NO_NERD_FONTS`** | Set to any value to use standard ASCII characters instead of Nerd Font icons. |
| **`ATELIER_HOSTNAME`** | Override the detected system hostname for host-specific configuration. |
| **`ATELIER_CLIENT_ID`** | Used for session recovery on remote machines (see [Remote Work](#remote-work)). |
| **`XDG_CONFIG_HOME`** | Custom location for configuration files (defaults to `~/.config`). |

## Usage

### Picker UI

Running `atelier-go` without arguments (or using the `ui` command) opens an interactive TUI. The UI displays both your configured projects and your most frequent `zoxide` directories, including their shortened paths for easy identification.

**Note:** Configured projects are always prioritized and shown at the top of the list, even when filtering.

#### Icons

By default, Atelier Go uses Nerd Font icons for folders, projects, and search. If you are not using a Nerd Font, you can disable these icons by setting the `NO_NERD_FONTS` environment variable:

```bash
export NO_NERD_FONTS=1
```

#### Triggers

Atelier Go uses a unified trigger system for all locations:

| Action | Key | Description |
| :--- | :--- | :--- |
| **Select** | `Enter` / `Tab` | Drill into the action menu for the selected location.* |
| **Fast Select** | `Alt-Enter` | Instantly launch the **Default Action**. |

*\*If a location has no configured actions (global or project-specific), `Enter` will instantly launch the default action (Shell).*

The **Default Action** is the first action in the list. By default, this is the first project-specific action or the first global action. If `shell-default` is set to `true`, then "Shell" becomes the default action.

#### Filters

*   **`atelier-go ui --projects`**: Filter to just your defined projects.
*   **`atelier-go ui --zoxide`**: Filter to just your `zoxide` directories.

### Sessions

If you prefer using the CLI over the interactive UI, you can manage your persistent `zmx` sessions directly:

*   **List sessions**: `atelier-go sessions list`
*   **Kill a session**: `atelier-go sessions kill <name>`
*   **Attach to a project**: `atelier-go sessions attach -p my-project`
*   **Run a specific action**: `atelier-go sessions attach -p my-project -a "Run Server"`
*   **Jump into a folder**: `atelier-go sessions attach -f ~/some/path`

You can use the reserved `--action Shell` to bypass a project's default action and just open a shell.

### Locations

To just see a table of everything Atelier Go has discovered, use the `locations` command:

```bash
# List everything
atelier-go locations

# List only projects
atelier-go locations --projects
```

## Remote Work

Atelier Go is designed to make working on remote machines feel seamless. By combining it with `autossh` and its built-in session recovery, you can maintain persistent remote connections that survive network drops.

### 1. Remote Server Setup

Ensure `atelier-go` is installed on your remote machine (e.g., at `~/.local/bin/atelier-go`).

### 2. SSH Configuration

Add a host entry to your local `~/.ssh/config`. Using `ControlMaster` and `ServerAlive` settings helps maintain a stable connection.

```ssh
Host ag
  HostName workstation
  User username
  ControlMaster auto
  ControlPath ~/.ssh/cm-%C
  ControlPersist 10m
  ServerAliveInterval 60
  ServerAliveCountMax 3
  RequestTTY yes
  LogLevel QUIET
```

### 3. Session Recovery (Recommended)

Atelier Go supports automatic session recovery via the `--client-id` flag. When provided, it tracks the active session and can automatically re-attach if the connection is interrupted and restarted by `autossh`.

Add a shell function to your local configuration (e.g., `~/.zshrc` or `~/.bashrc`):

```bash
agr() {
    # Generate a unique ID for this terminal tab if not already set
    export ATELIER_CLIENT_ID="${ATELIER_CLIENT_ID:-$(uuidgen | cut -d'-' -f1)}"
    
    # Use autossh to maintain the connection and pass the client ID
    autossh -M 0 -q -t ag -- "/home/username/.local/bin/atelier-go --client-id=$ATELIER_CLIENT_ID"
}
```

*   **`autossh`**: Automatically restarts the SSH session if it drops.
*   **`--client-id`**: Tells the remote Atelier Go which session to recover. Using a unique ID per terminal tab allows multiple concurrent remote sessions to recover independently.
*   **`-t`**: Forces a PTY allocation, which is required for the interactive UI.

Now, if your laptop goes to sleep or your connection drops, `autossh` will reconnect and Atelier Go will instantly drop you back into your active session.

### 4. Basic Remote Usage (No Recovery)

If you don't need recovery, a simple alias works:

```bash
alias agr='ssh -t ag -- /home/username/.local/bin/atelier-go'
```


## Inspiration and Prior Art
These are the projects that inspired this.

- [Ghostty Terminal](https://github.com/ghostty-org/ghostty): Using this terminal inspired be to come up with a more native workflow.
- [ZMX](https://github.com/neurosnap/zmx): Used for managing sessions on Mac and Linux.
- [Shpool](https://github.com/shell-pool/shpool): Another session manager. I used this originally, but switched to ZMX for cross-platform support.
- [Sesh](https://github.com/joshmedeski/sesh): I've never actually used this, but saw some videos of it and it helped inspire how this works.

