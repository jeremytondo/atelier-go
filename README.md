# Atelier Go

Atelier Go is a workspace launcher and session manager that focuses on supporting a native terminal experience free of multiplexers like Tmux.
It allows you to quickly jump into projects and apps, making it easier to work with native terminal windows, tabs, and splits.

## Table of Contents

- [About](#about)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
  - [Projects](#projects)
  - [Host-Specific Projects](#host-specific-projects)
- [Usage](#usage)
  - [Interactive UI](#interactive-ui)
  - [Sessions](#sessions)
  - [Locations](#locations)
- [Remote Work: Environment Bootstrapping](#remote-work-environment-bootstrapping)

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

You can define global settings like your preferred editor at the top of `config.yaml`.

```yaml
editor: "nvim"
```

*   **`editor`**: The command used to open folders (e.g., `nvim`, `vim`, `code`). If not set, it defaults to the `$EDITOR` environment variable, then `vim`.

### Projects

You can define projects by creating a `config.yaml` file in `~/.config/atelier-go/`.

**Example: `~/.config/atelier-go/config.yaml`**

```yaml
projects:
  - name: "My Application"
    path: "~/dev/my-app"
    actions:
      - name: "Run Server"
        command: "npm start"
      - name: "Build"
        command: "make build"
```

*   **`name`**: The display name shown in the UI.
*   **`path`**: The directory to jump into (supports `~` expansion).
*   **`actions`**: Custom commands you can run.

### Host-Specific Projects

If you work across multiple machines, you can define projects that only show up on a specific host. Create a YAML file named after the host in the same directory:

`~/.config/atelier-go/<hostname>.yaml`

Settings in the host-specific file will be merged with the global `config.yaml`.

To see what hostname Atelier Go is using for your current host, run:

```bash
atelier-go hostname
```

## Usage

### Picker UI

Running `atelier-go` without arguments (or using the `ui` command) opens an interactive `fzf` window.

#### Triggers

Atelier Go uses a dual-trigger system to launch different actions depending on the location type:

| Location Type | Primary (`Enter`) | Secondary (`Alt-Enter`) |
| :--- | :--- | :--- |
| **Folder** (Zoxide) | Open Shell | Open Editor |
| **Project** (Configured) | Run Default Action* | Open Action Menu |

*\*Default Action is the first action defined in the project configuration. If no actions are defined, it falls back to a shell.*

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

You can even use the reserved `--action Shell` to bypass a project's default action and just open a shell.

### Locations

To just see a table of everything Atelier Go has discovered, use the `locations` command:

```bash
# List everything
atelier-go locations

# List only projects
atelier-go locations --projects
```

## Remote Work

Atelier Go is designed to make working on remote machines feel native. One interesting way
to set this up is by running it via ssh. This allows you to easily work with projects
on remote machines.

### 1. Remote Server Setup

Ensure `atelier-go` is installed on your remote machine (e.g., at `~/.local/bin/atelier-go`).

### 2. SSH Configuration

Add a host entry to your local `~/.ssh/config`.

```ssh
Host ag
  HostName workstation
  User username
  ControlMaster auto
  ControlPath ~/.ssh/cm-%C
  ControlPersist 10m
  ServerAliveInterval 60
  RequestTTY yes # Required for the interactive UI
  LogLevel QUIET
```

### 3. Local Alias

Add the following alias to your local shell configuration (e.g., `~/.zshrc` or `~/.bashrc`):

```bash
alias agr='ssh -t ag -- /home/username/.local/bin/atelier-go'
```

*   **`-t`**: Forces a PTY allocation, which is required for the `fzf` UI.
*   **`--`**: Ensures all following flags are passed to `atelier-go` instead of `ssh`.

Now you can use the remote version of Atelier Go seamlessly:

```bash
# Launch the interactive UI
agr

# List remote sessions
agr sessions list

# Start with specific UI filters
agr ui --projects
```

## Inspiration and Prior Art
These are the projects that inspired this.

- [Ghostty Terminal](https://github.com/ghostty-org/ghostty): Using this terminal inspired be to come up with a more native workflow.
- [ZMX](https://github.com/neurosnap/zmx): Used for managing sessions on Mac and Linux.
- [Shpool](https://github.com/shell-pool/shpool): Another session manager. I used this originally, but switched to ZMX for cross-platform support.
- [Sesh](https://github.com/joshmedeski/sesh): I've never actually used this, but saw some videos of it and it helped inspire how this works.

