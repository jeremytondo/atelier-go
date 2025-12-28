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

### Projects

You can define projects by creating TOML files in `~/.config/atelier-go/projects/`. Each file represents a separate project.

**Example: `~/.config/atelier-go/projects/my-app.toml`**

```toml
name = "My Application"
path = "~/dev/my-app"

[[actions]]
  name = "Run Server"
  command = "npm start"

[[actions]]
  name = "Build"
  command = "make build"
```

*   **`name`**: The display name shown in the UI.
*   **`path`**: The directory to jump into (supports `~` expansion).
*   **`actions`**: Custom commands you can run.

### Host-Specific Projects

If you work across multiple machines, you can define projects that only show up on a specific host. Place them in a subdirectory named after the host:
`~/.config/atelier-go/projects/<hostname>/my-app.toml`

To see what hostname Atelier Go is using for your current host, run:
```bash
atelier-go hostname
```

## Usage

### Picker UI

Running `atelier-go` without arguments (or using the `ui` command) opens an interactive `fzf` window.

*   **Default View**: Shows everything (Projects + Zoxide).
*   **`atelier-go ui --projects`**: Filter to just your defined projects.
*   **`atelier-go ui --zoxide`**: Filter to just your `zoxide` directories.

### Sessions

You can also manage your persistent `zmx` sessions directly from the CLI:

*   **List sessions**: `atelier-go sessions list`
*   **Kill a session**: `atelier-go sessions kill <name>`
*   **Manual attach**: `atelier-go sessions attach <name> <path> [command]`

### Locations

To just see a table of everything Atelier Go has discovered, use the `locations` command:

```bash
# List everything
atelier-go locations

# List only projects
atelier-go locations --projects
```

## Remote Work

Atelier Go tries to make working on remote machines as easy as working locally.
This is supported by installing Atelier Go on the remote machine and then running
it over ssh. There are a few other neat tricks that can be done as well that I learned
from using [zmx](https://github.com/neurosnap/zmx).

Here's an example ssh config:

```bash
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
  RemoteCommand /home/username/.local/bin/atelier-go

```

Then, if you want to run Atelier Go on the remote host just do this:

```bash
ssh ag
```

## Inspiration and Prior Art
These are the projects that inspired this.

- [Ghostty Terminal](https://github.com/ghostty-org/ghostty): Using this terminal inspired be to come up with a more native workflow.
- [ZMX](https://github.com/neurosnap/zmx): Used for managing sessions on Mac and Linux.
- [Shpool](https://github.com/shell-pool/shpool): Another session manager. I used this originally, but switched to ZMX for cross-platform support.
- [Sesh](https://github.com/joshmedeski/sesh): I've never actually used this, but saw some videos of it and it helped inspire how this works.

