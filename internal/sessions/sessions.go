// Package sessions handles workspace session management.
package sessions

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// Session represents a running workspace session.
type Session struct {
	ID   string
	Path string
}

// Manager handles interaction with the zmx session manager.
type Manager struct{}

// NewManager creates a new session manager.
func NewManager() *Manager {
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

// List returns a list of active sessions (stub implementation as zmx list format is not specified).
// Assuming 'zmx list' returns "ID\tPath" or similar.
func (m *Manager) List() ([]Session, error) {
	// TODO: Implement actual parsing of `zmx list` output if available.
	return nil, nil
}

// Kill terminates a session.
func (m *Manager) Kill(name string) error {
	cmd := exec.Command("zmx", "kill", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to kill session %s: %w", name, err)
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
