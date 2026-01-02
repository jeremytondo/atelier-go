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

	tea "github.com/charmbracelet/bubbletea"
)

// Run executes the interactive UI.
// It fetches locations using the provided manager, prompts the user, and attaches to a session.
func Run(ctx context.Context, mgr *locations.Manager, cfg *config.Config) error {
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
	fmt.Printf("Attaching to %ssession '%s' in %s\n", statusPrefix, result.Name, result.Path)
	if err := sessionManager.Attach(result.Name, result.Path, result.Command...); err != nil {

		return fmt.Errorf("error attaching to session: %w", err)
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
