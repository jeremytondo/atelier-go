package zmx

import (
	"fmt"
	"os"
	"os/exec"
)

// Attach connects to an existing zmx session or creates a new one with the given name.
// It sets the working directory to dir for new sessions.
// It connects the current process's Stdin, Stdout, and Stderr to the zmx command.
func Attach(name string, dir string) error {
	// zmx attach <name> [command...]
	// We omit the command to let zmx default to the user's shell.
	cmd := exec.Command("zmx", "attach", name)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("zmx session ended with error: %w", err)
	}

	return nil
}
