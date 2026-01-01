// Package ui provides the interactive user interface.
package ui

import (
	"atelier-go/internal/config"
	"atelier-go/internal/env"
	"atelier-go/internal/locations"
	"atelier-go/internal/sessions"
	"atelier-go/internal/utils"
	"context"
	"errors"
	"fmt"
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

// runSelection executes the selection logic.
func runSelection(locs []locations.Location, cfg *config.Config) (*sessions.Target, error) {
	displayMap := make(map[string]locations.Location)
	var choices []string

	for _, loc := range locs {
		label := FormatLocation(loc)
		choices = append(choices, label)
		// Use stripped label as key for robustness
		displayMap[StripANSI(label)] = loc
	}

	selection, key, err := Select(choices, "Select Project (Alt-Enter for Secondary)", "Project ➜ ", []string{"alt-enter"})
	if err != nil {
		if errors.Is(err, ErrCancelled) {
			return nil, nil
		}
		return nil, err
	}

	item, ok := displayMap[StripANSI(selection)]
	if !ok {
		// Fallback: Try exact match
		if i, okExact := displayMap[selection]; okExact {
			item = i
		} else {
			return nil, fmt.Errorf("invalid selection: %q", selection)
		}
	}

	secondary := key == "alt-enter"
	shell := env.DetectShell()
	sessionManager := sessions.NewManager()

	if secondary {
		if item.Source == "Project" {
			return showActionMenu(item, shell, cfg.GetEditor(), sessionManager)
		}
		// Folder Secondary -> Editor
		return sessionManager.Resolve(item, "editor", shell, cfg.GetEditor())
	}

	// Primary action
	return sessionManager.Resolve(item, "", shell, cfg.GetEditor())
}

func showActionMenu(item locations.Location, shell string, editor string, sessionManager *sessions.Manager) (*sessions.Target, error) {
	if len(item.Actions) == 0 {
		// If no actions, just return shell
		return sessionManager.Resolve(item, "", shell, editor)
	}

	type actEntry struct {
		name string
	}
	actionMap := make(map[string]actEntry)
	var actionChoices []string

	// Add actions
	for i, act := range item.Actions {
		label := act.Name
		if i == 0 {
			label = fmt.Sprintf("%s (Default)", label)
		}
		actionChoices = append(actionChoices, label)
		actionMap[label] = actEntry{name: act.Name}
	}

	// Add Shell option
	shellLabel := "Shell"
	actionChoices = append(actionChoices, shellLabel)
	actionMap[shellLabel] = actEntry{name: "shell"}

	actSelection, _, err := Select(actionChoices, fmt.Sprintf("Select Action for %s", item.Name), "Action ➜ ", nil)
	if err != nil {
		if errors.Is(err, ErrCancelled) {
			return nil, nil
		}
		return nil, err
	}

	if entry, ok := actionMap[actSelection]; ok {
		return sessionManager.Resolve(item, entry.name, shell, editor)
	}

	return sessionManager.Resolve(item, "", shell, editor)
}
