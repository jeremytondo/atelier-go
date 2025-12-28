// Package env handles environment bootstrapping and management.
package env

import (
	"fmt"
	"os"
	"strings"
	"syscall"
)

// Bootstrap ensures the application has a complete environment.
// It detects restricted SSH contexts and "promotes" the process by re-executing
// it inside an interactive login shell.
func Bootstrap() {
	// 1. Check if we need to bootstrap
	if os.Getenv("SSH_CONNECTION") == "" {
		return // Not remote
	}
	if os.Getenv("ATELIER_PROMOTED") == "1" {
		return // Already promoted
	}

	// 2. Identify the Login Shell
	// Try the SHELL env var, upgrade from /bin/sh if possible
	shell := os.Getenv("SHELL")
	if shell == "" || shell == "/bin/sh" {
		if _, err := os.Stat("/bin/zsh"); err == nil {
			shell = "/bin/zsh"
		} else if _, err := os.Stat("/bin/bash"); err == nil {
			shell = "/bin/bash"
		} else {
			shell = "/bin/sh"
		}
	}

	// 3. Identify Self and Arguments
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to determine executable path: %v\n", err)
		return
	}
	args := os.Args[1:]

	// 4. Construct the Re-Exec Command
	// We construct a command string that:
	// - Exports the sentinel variable
	// - Execs the original binary with its arguments
	// The 'exec' is important to replace the shell process with the app.

	cmdBuilder := strings.Builder{}
	cmdBuilder.WriteString("export ATELIER_PROMOTED=1; exec ")
	cmdBuilder.WriteString(quote(exe))
	for _, arg := range args {
		cmdBuilder.WriteString(" ")
		cmdBuilder.WriteString(quote(arg))
	}

	// 5. Promote via syscall.Exec
	// We replace the current process with the shell process.
	// The shell will be interactive (-i) and login (-l) to load all configs.
	shellArgs := []string{shell, "-l", "-i", "-c", cmdBuilder.String()}

	// syscall.Exec requires the full environment array
	env := os.Environ()

	if err := syscall.Exec(shell, shellArgs, env); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to promote shell: %v\n", err)
		// Fall through to normal execution if promotion fails
	}
}

// quote escapes a string for shell usage (single quoting)
func quote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
