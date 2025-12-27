// Package zmx handles the interaction with the zmx session manager.
package zmx

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// Manager implements the session.Manager interface for zmx.
type Manager struct{}

// New creates a new zmx session manager.
func New() *Manager {
	return &Manager{}
}

// Attach connects to an existing zmx session or creates a new one with the given name.
func (m *Manager) Attach(name string, dir string, args ...string) error {
	cmdArgs := []string{"attach", name}
	if len(args) > 0 {
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

// Sanitize cleans a string to be used as a session name component.
func Sanitize(s string) string {
	s = strings.ToLower(s)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}
