// Package sessions handles workspace session management.
package sessions

import (
	"atelier-go/internal/config"
	"atelier-go/internal/env"
	"atelier-go/internal/locations"
	"atelier-go/internal/utils"
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Session represents a running workspace session.
type Session struct {
	ID   string
	Path string
}

// Target represents a resolved session target ready for attachment.
type Target struct {
	Name    string
	Path    string
	Command []string
}

// Manager handles interaction with the zmx session manager.
type Manager struct{}

// NewManager creates a new session manager.
func NewManager() *Manager {
	return &Manager{}
}

// Resolve converts a location and optional action into a concrete Target.
func (m *Manager) Resolve(loc locations.Location, actionName string, shell string, editor string) (*Target, error) {
	sanitizedAction := utils.Sanitize(actionName)

	// 1. If an actionName is provided, try to find it in loc.Actions first.
	// This respects custom "shell" or "editor" actions.
	if actionName != "" {
		for _, act := range loc.Actions {
			if utils.Sanitize(act.Name) == sanitizedAction {
				return m.resolveAction(loc, act, shell)
			}
		}

		// 2. Fallback to built-in behaviors if not found in loc.Actions
		if sanitizedAction == "editor" {
			return &Target{
				Name:    utils.Sanitize(loc.Name) + ":editor",
				Path:    loc.Path,
				Command: env.BuildInteractiveWrapper(shell, editor+" ."),
			}, nil
		}

		if sanitizedAction == "shell" {
			return &Target{
				Name:    utils.Sanitize(loc.Name),
				Path:    loc.Path,
				Command: env.BuildInteractiveWrapper(shell, ""),
			}, nil
		}

		return nil, fmt.Errorf("action %q not found for %q", actionName, loc.Name)
	}

	// 3. No actionName provided. Use default action (first one) if it exists.
	if len(loc.Actions) > 0 {
		return m.resolveAction(loc, loc.Actions[0], shell)
	}

	// 4. No actions exist, open a shell.
	return &Target{
		Name:    utils.Sanitize(loc.Name),
		Path:    loc.Path,
		Command: env.BuildInteractiveWrapper(shell, ""),
	}, nil
}

// resolveAction creates a Target from a specific action.
func (m *Manager) resolveAction(loc locations.Location, act config.Action, shell string) (*Target, error) {
	name := fmt.Sprintf("%s:%s", utils.Sanitize(loc.Name), utils.Sanitize(act.Name))
	// If the action is "Shell", don't add the suffix for consistency
	if utils.Sanitize(act.Name) == "shell" {
		name = utils.Sanitize(loc.Name)
	}
	return &Target{
		Name:    name,
		Path:    loc.Path,
		Command: env.BuildInteractiveWrapper(shell, act.Command),
	}, nil
}

// Attach connects to an existing zmx session or creates a new one with the given name.
func (m *Manager) Attach(name string, dir string, args ...string) error {
	utils.SetTerminalTitle(name)

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
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return nil
		}
		return fmt.Errorf("zmx session ended with error: %w", err)
	}

	return nil
}

// List returns a list of active sessions.
// It assumes 'zmx list' returns output where each line is "ID" or "ID\tPath".
func (m *Manager) List() ([]Session, error) {
	cmd := exec.Command("zmx", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	var sessions []Session
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "\t", 2)
		id := parts[0]
		// Handle session_name= prefix if present
		id = strings.TrimPrefix(id, "session_name=")

		sess := Session{ID: id}
		if len(parts) > 1 {
			sess.Path = parts[1]
		}
		sessions = append(sessions, sess)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse session list: %w", err)
	}

	return sessions, nil
}

// Kill terminates a session.
func (m *Manager) Kill(name string) error {
	cmd := exec.Command("zmx", "kill", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to kill session %s: %w", name, err)
	}
	return nil
}

// SessionExists checks if a session with the given ID is currently running.
func (m *Manager) SessionExists(sessionID string) bool {
	sessions, err := m.List()
	if err != nil {
		return false
	}
	for _, s := range sessions {
		if s.ID == sessionID {
			return true
		}
	}
	return false
}

// SaveState writes the current session ID to the state file for the given client.
func SaveState(clientID, sessionID string) error {
	dir, err := utils.GetStateDir()
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, clientID), []byte(sessionID), 0600)
}

// LoadState reads the session ID from the state file for the given client.
func LoadState(clientID string) (string, error) {
	dir, err := utils.GetStateDir()
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(filepath.Join(dir, clientID))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(content), nil
}

// ClearState removes the state file for the given client.
func ClearState(clientID string) error {
	dir, err := utils.GetStateDir()
	if err != nil {
		return err
	}
	err = os.Remove(filepath.Join(dir, clientID))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// PrintTable formats and prints the sessions to the provided writer in a table format.
func (m *Manager) PrintTable(w io.Writer, sessions []Session) error {
	if len(sessions) == 0 {
		if _, err := fmt.Fprintln(w, "No active sessions found."); err != nil {
			return fmt.Errorf("error writing output: %w", err)
		}
		return nil
	}

	headers := []string{"ID", "PATH"}
	var rows [][]string

	for _, s := range sessions {
		rows = append(rows, []string{s.ID, s.Path})
	}

	return utils.RenderTable(w, headers, rows)
}
