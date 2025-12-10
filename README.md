# Atelier Go

## About

Atelier Go is a CLI tool that helps streamline development in the terminal, especially when you're working on remote machines.

I was inspired by the Ghostty terminals dedication to native UI. The way they integrate with the native UI—windows, tabs, splits—just feels right. It's hard to put into words, but it makes everything feel cohesive with the rest of the operating system.

The problem is, that "native" feeling often breaks down when you're doing most of your development on remote workstations, which is my usual setup. Those native terminal features become less useful because every time you open a new window or tab, you have to SSH back into your remote machine and get your development environment set up again. Tools like Tmux or Zellij are the standard solution here, and I actually like them a lot. But I really wanted something that let me keep using those nice native UI features of Ghostty.

That's where Atelier Go comes in. The idea is pretty simple: it's a launcher and session manager. It lets you quickly jump into specific locations, projects, or active sessions on your remote workstation directly from a new local terminal window, tab, or split. This makes it much more practical to use your terminal's native UI features, rather than relying solely on a multiplexer like Tmux.

Right now, the main challenge is that it's still a two-step process: open a window, tab or split then launch your remote session. Atelier Go makes that second step much faster, but it's still a distinct action. Ideally, I'd love to see custom keybinds in Ghostty that would allow chained actions. This would allow you to open a window and then run a custom script to setup the environment all in one go. That would get us really close to a Tmux-like workflow, but with all the benefits of native UI.

## Quick Start

Here’s how to get Atelier Go up and running quickly:

1.  **Download the binary:**
    Grab the latest release for your operating system from the [GitHub Releases page](https://github.com/jeremytondo/atelier-go/releases).
    *   For Linux/macOS, you'll typically download a `.tar.gz` file.
    *   For Windows, you'll usually download a `.zip` file.

2.  **Verify installation:**
    Open a new terminal and run:
    ```bash
    atelier-go --version
    ```
    You should see the version number printed.

3.  **Start using Atelier Go:**
    Once installed, you can start configuring your remote projects and sessions.
    ```bash
    atelier-go server start # Start the server in the background
    atelier-go client # Connect to the server to select a session or project
    ```
    Refer to the Documentation section for detailed usage.

## Documentation

Atelier Go is a unified tool for managing remote development environments.

It consists of two parts:
1. A Server that runs on your development machine/host.
2. A Client that you run to connect, manage, and attach to sessions.

It integrates with 'shpool' for persistent sessions and 'zoxide' for smart path navigation.

### Global Flags

*   `-h, --help`: help for atelier-go
*   `--host string`: Host to connect to (client) or bind to (server)
*   `--port int`: Port to connect to or listen on (default 9001)
*   `-v, --version`: Print the version number

### Commands

#### `atelier-go client`

The client command connects to the running Atelier server.

It presents an interactive list of:
- Active 'shpool' sessions
- Defined projects
- Frequent directories (via zoxide)

You can filter this list using flags. Selecting an item will either attach
to an existing session or start a new one in that location.

**Usage:**
```
atelier-go client [flags]
atelier-go client [command]
```

**Examples:**
```bash
# Default: Show sessions, projects, and frequent paths
atelier-go client

# Show only active sessions
atelier-go client --sessions

# Show ALL directories (can be slow)
atelier-go client --all
```

**Flags:**
*   `-a, --all`: Show all directories in home
*   `-h, --help`: help for client
*   `-s, --sessions`: Show open sessions only

##### `atelier-go client login`

Save the authentication token

**Usage:**
```
atelier-go client login [token] [flags]
```

**Flags:**
*   `-h, --help`: help for login

#### `atelier-go completion`

Generate the autocompletion script for the specified shell.

#### `atelier-go server`

Starts the Atelier server process in the current terminal window.
This is useful for debugging or running in a container.
For normal usage, consider using 'server start' to run in the background.

**Usage:**
```
atelier-go server [flags]
atelier-go server [command]
```

**Examples:**
```bash
atelier-go server --port 9005
```

**Flags:**
*   `-h, --help`: help for server

##### `atelier-go server install`

Generates a systemd user service file and installs it.
This allows the Atelier server to start automatically when you log in.
Files are installed to ~/.local/share/systemd/user/.

**Usage:**
```
atelier-go server install [flags]
```

**Examples:**
```bash
atelier-go server install
```

**Flags:**
*   `-h, --help`: help for install

##### `atelier-go server start`

Starts the Atelier server as a detached background process.
It writes a PID file to ensure only one instance runs at a time.
Use 'server stop' to shut it down.

**Usage:**
```
atelier-go server start [flags]
```

**Examples:**
```bash
atelier-go server start
```

**Flags:**
*   `-h, --help`: help for start

##### `atelier-go server status`

Sends a health check request to the running server.
Verifies that the server process is running and responding to HTTP requests.

**Usage:**
```
atelier-go server status [flags]
```

**Examples:**
```bash
atelier-go server status
```

**Flags:**
*   `-h, --help`: help for status

##### `atelier-go server stop`

Stops the currently running background server instance
identified by the PID file.

**Usage:**
```
atelier-go server stop [flags]
```

**Examples:**
```bash
atelier-go server stop
```

**Flags:**
*   `-h, --help`: help for stop

##### `atelier-go server token`

Retrieves and prints the current authentication token.
This is useful if you need to manually configure a client or script
authentication.

**Usage:**
```
atelier-go server token [flags]
```

**Examples:**
```bash
atelier-go server token
```

**Flags:**
*   `-h, --help`: help for token
