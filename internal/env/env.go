// Package env handles environment bootstrapping and management.
package env

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Bootstrap ensures the application has a complete environment,
// particularly when running in restricted contexts like SSH RemoteCommand.
func Bootstrap() {
	// Only run if we are in an SSH connection
	if os.Getenv("SSH_CONNECTION") == "" {
		return
	}

	// Detect the "Real" Login Shell
	// In restricted SSH environments, $SHELL often points to /bin/sh or is empty.
	// We try to find a better interactive shell if possible.
	shell := os.Getenv("SHELL")
	if shell == "" || shell == "/bin/sh" {
		// Try to upgrade to a more capable shell for harvesting
		if _, err := os.Stat("/bin/zsh"); err == nil {
			shell = "/bin/zsh"
		} else if _, err := os.Stat("/bin/bash"); err == nil {
			shell = "/bin/bash"
		} else {
			// Fallback
			shell = "/bin/sh"
		}
	}

	// We harvest the PATH from an interactive login shell.
	// We use both -l (login) and -i (interactive) to ensure all user config
	// files (.bashrc, .profile, .zshrc etc) are loaded.
	// We echo the PATH variable to stdout.
	cmd := exec.Command(shell, "-l", "-i", "-c", "echo $PATH")

	// Capture output
	var out bytes.Buffer
	cmd.Stdout = &out
	// We ignore stderr as interactive shells might print banners/motd there
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		// If we fail to harvest, we just proceed with the current environment.
		// Logging to stderr might be helpful but we don't want to break the app start.
		fmt.Fprintf(os.Stderr, "Warning: failed to harvest environment from %s: %v\n", shell, err)
		return
	}

	// Parse the output.
	// The shell might print other things (like pre-command prompts if not carefully handled),
	// but generally 'echo $PATH' should be the last thing printed to stdout.
	// However, interactive shells are noisy.
	// To be safer, we could have used a marker, but for now let's trust that
	// standard output from a non-tty connected command (which cmd.Run does)
	// typically only contains what we asked for, unless the user's rc files print to stdout.

	// Let's trim whitespace
	harvestedPath := strings.TrimSpace(out.String())

	// Basic validation: A PATH should look like a list of paths
	if harvestedPath != "" {
		// If the output contains multiple lines, the user's RC files are printing to stdout.
		// We'll take the last non-empty line as the likely PATH.
		lines := strings.Split(harvestedPath, "\n")
		for i := len(lines) - 1; i >= 0; i-- {
			line := strings.TrimSpace(lines[i])
			if line != "" && strings.Contains(line, "/") {
				harvestedPath = line
				break
			}
		}

		if err := os.Setenv("PATH", harvestedPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to set PATH: %v\n", err)
		}
	}
}
