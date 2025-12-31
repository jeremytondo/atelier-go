// Package sessions handles workspace session management.
package sessions

import (
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
	"regexp"
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
	if loc.Source == "Project" {
		// If actionName is provided, look it up
		if actionName != "" {
			for _, act := range loc.Actions {
				if Sanitize(act.Name) == Sanitize(actionName) {
					return &Target{
						Name:    fmt.Sprintf("%s:%s", Sanitize(loc.Name), Sanitize(act.Name)),
						Path:    loc.Path,
						Command: env.BuildInteractiveWrapper(shell, act.Command),
					}, nil
				}
			}
			return nil, fmt.Errorf("action %q not found in project %q", actionName, loc.Name)
		}

		// No actionName, use default action (first one) or shell
		if len(loc.Actions) > 0 {
			act := loc.Actions[0]
			return &Target{
				Name:    fmt.Sprintf("%s:%s", Sanitize(loc.Name), Sanitize(act.Name)),
				Path:    loc.Path,
				Command: env.BuildInteractiveWrapper(shell, act.Command),
			}, nil
		}
	} else {
		// Zoxide / Folder
		if actionName == "editor" {
			return &Target{
				Name:    Sanitize(loc.Name) + ":editor",
				Path:    loc.Path,
				Command: env.BuildInteractiveWrapper(shell, editor+" ."),
			}, nil
		}
	}

	// Default: Open Shell
	return &Target{
		Name:    Sanitize(loc.Name),
		Path:    loc.Path,
		Command: env.BuildInteractiveWrapper(shell, ""),
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
		sess := Session{ID: parts[0]}
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

var sanitizeRegex = regexp.MustCompile(`[^a-z0-9]+`)

// Sanitize cleans a string to be used as a session name component.
func Sanitize(s string) string {
	s = strings.ToLower(s)
	s = sanitizeRegex.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
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
