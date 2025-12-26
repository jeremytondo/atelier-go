package zmx

import (
	"fmt"
	"os"
	"os/exec"
)

// Attach connects to an existing zmx session or creates a new one with the given name.
// It sets the working directory to dir for new sessions.
// It connects the current process's Stdin, Stdout, and Stderr to the zmx command.
func Attach(name string, dir string, args ...string) error {
	// zmx attach <name> [command...]
	// We omit the command to let zmx default to the user's shell.
	cmdArgs := []string{"attach", name}
	if len(args) > 0 {
		// If command arguments are provided (e.g. "nvim ."), zmx expects them as trailing arguments
		// However, zmx attach signature is `zmx attach <session_name> [command...]`
		// Wait, I need to double check if my assumption about zmx attach is correct.
		// The original code was `exec.Command("zmx", "attach", name)`.
		// If I add args... it should just be append.
		cmdArgs = append(cmdArgs, args...)
	}

	cmd := exec.Command("zmx", cmdArgs...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("zmx session ended with error: %w", err)
	}

	return nil
}
