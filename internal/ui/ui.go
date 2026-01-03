// Package ui provides the interactive user interface.
package ui

import (
	"atelier-go/internal/config"
	"atelier-go/internal/env"
	"atelier-go/internal/locations"
	"atelier-go/internal/sessions"
	"atelier-go/internal/utils"
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Run executes the interactive UI.
// It fetches locations using the provided manager, prompts the user, and attaches to a session.
func Run(ctx context.Context, mgr *locations.Manager, cfg *config.Config, clientID string) error {
	// Apply custom theme from config
	ApplyTheme(cfg.Theme)

	// Try to recover session if client ID is provided
	if clientID != "" {
		if recovered := tryRecover(clientID); recovered {
			return nil
		}
	}

	// Fetch locations
	locs, err := mgr.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch locations: %w", err)
	}

	if len(locs) == 0 {
		fmt.Println("No projects or recent directories found.")
		return nil
	}

	// Interactive selection
	result, err := runSelection(locs, cfg)
	if err != nil {
		return err
	}
	if result == nil {
		// User cancelled
		return nil
	}

	// Attach to session
	sessionManager := sessions.NewManager()
	statusPrefix := ""
	if utils.IsSSH() {
		statusPrefix = utils.IconSSH + " "
	}

	// Save state before attaching
	if clientID != "" {
		if err := sessions.SaveState(clientID, result.Name); err != nil {
			fmt.Printf("Warning: failed to save session state: %v\n", err)
		}
	}

	fmt.Printf("Attaching to %ssession '%s' in %s\n", statusPrefix, result.Name, result.Path)
	if err := sessionManager.Attach(result.Name, result.Path, result.Command...); err != nil {
		return fmt.Errorf("error attaching to session: %w", err)
	}

	// Clear state after clean exit
	if clientID != "" {
		if err := sessions.ClearState(clientID); err != nil {
			fmt.Printf("Warning: failed to clear session state: %v\n", err)
		}
	}

	return nil
}

// runSelection executes the TUI and returns a session target
func runSelection(locs []locations.Location, cfg *config.Config) (*sessions.Target, error) {
	model := NewModel(locs)

	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("TUI error: %w", err)
	}

	m, ok := finalModel.(*Model)
	if !ok {
		return nil, fmt.Errorf("unexpected model type")
	}

	if m.Result.Canceled || m.Result.Location == nil {
		return nil, nil // User cancelled
	}

	// Resolve selection to session target
	shell := env.DetectShell()
	sessionManager := sessions.NewManager()

	actionName := ""
	if m.Result.Action != nil {
		actionName = m.Result.Action.Name
	}

	return sessionManager.Resolve(*m.Result.Location, actionName, shell, cfg.GetEditor())
}

// tryRecover attempts to re-attach to a previously active session for the client.
func tryRecover(clientID string) bool {
	sessionID, err := sessions.LoadState(clientID)
	if err != nil || sessionID == "" {
		return false
	}

	mgr := sessions.NewManager()
	if !mgr.SessionExists(sessionID) {
		// Session no longer exists, clean up stale state
		_ = sessions.ClearState(clientID)
		return false
	}

	// Session exists, re-attach directly
	fmt.Printf("%s Recovering session '%s'...\n", utils.IconSSH, sessionID)
	if err := mgr.Attach(sessionID, ""); err != nil {
		fmt.Fprintf(os.Stderr, "error during recovery: %v\n", err)
		return false
	}

	// Successfully re-attached, and clean exit
	_ = sessions.ClearState(clientID)
	return true
}
