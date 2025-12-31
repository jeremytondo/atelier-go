// Package ui provides the interactive user interface.
package ui

import (
	"atelier-go/internal/config"
	"atelier-go/internal/env"
	"atelier-go/internal/locations"
	"atelier-go/internal/sessions"
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
	fmt.Printf("Attaching to session '%s' in %s\n", result.SessionName, result.Path)
	if err := sessionManager.Attach(result.SessionName, result.Path, result.CommandArgs...); err != nil {
		return fmt.Errorf("error attaching to session: %w", err)
	}

	return nil
}

// WorkflowResult represents the outcome of the user interaction.
type WorkflowResult struct {
	SessionName string
	Path        string
	CommandArgs []string
}

// runSelection executes the selection logic.
func runSelection(locs []locations.Location, cfg *config.Config) (*WorkflowResult, error) {
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

	// Logic Matrix:
	// Folder  + Primary   -> Shell
	// Folder  + Secondary -> Editor
	// Project + Primary   -> Default Action (or Shell)
	// Project + Secondary -> Action Menu

	if item.Source == "Project" {
		if secondary {
			return showActionMenu(item, shell)
		}
		// Primary action for Project: First defined action or shell
		if len(item.Actions) > 0 {
			act := item.Actions[0]
			return &WorkflowResult{
				SessionName: fmt.Sprintf("%s:%s", sessions.Sanitize(item.Name), sessions.Sanitize(act.Name)),
				Path:        item.Path,
				CommandArgs: env.BuildInteractiveWrapper(shell, act.Command),
			}, nil
		}
	} else {
		// Zoxide / Folder
		if secondary {
			editor := cfg.GetEditor()
			return &WorkflowResult{
				SessionName: sessions.Sanitize(item.Name) + ":editor",
				Path:        item.Path,
				CommandArgs: env.BuildInteractiveWrapper(shell, editor+" ."),
			}, nil
		}
	}

	// Default: Open Shell
	return &WorkflowResult{
		SessionName: sessions.Sanitize(item.Name),
		Path:        item.Path,
		CommandArgs: env.BuildInteractiveWrapper(shell, ""),
	}, nil
}

func showActionMenu(item locations.Location, shell string) (*WorkflowResult, error) {
	if len(item.Actions) == 0 {
		// If no actions, just return shell (or we could show a message)
		return &WorkflowResult{
			SessionName: sessions.Sanitize(item.Name),
			Path:        item.Path,
			CommandArgs: env.BuildInteractiveWrapper(shell, ""),
		}, nil
	}

	type actEntry struct {
		cmd  string
		name string
	}
	actionMap := make(map[string]actEntry)
	var actionChoices []string

	// Add Shell option
	shellLabel := "Shell (Default)"
	actionChoices = append(actionChoices, shellLabel)
	actionMap[shellLabel] = actEntry{cmd: "", name: ""}

	for _, act := range item.Actions {
		label := act.Name
		actionChoices = append(actionChoices, label)
		actionMap[label] = actEntry{cmd: act.Command, name: sessions.Sanitize(act.Name)}
	}

	actSelection, _, err := Select(actionChoices, fmt.Sprintf("Select Action for %s", item.Name), "Action ➜ ", nil)
	if err != nil {
		if errors.Is(err, ErrCancelled) {
			return nil, nil
		}
		return nil, err
	}

	sessionName := sessions.Sanitize(item.Name)
	var commandArgs []string

	if entry, ok := actionMap[actSelection]; ok {
		if entry.name != "" {
			sessionName = fmt.Sprintf("%s:%s", sessions.Sanitize(item.Name), entry.name)
		}
		if entry.cmd != "" {
			commandArgs = env.BuildInteractiveWrapper(shell, entry.cmd)
		}
	}

	if len(commandArgs) == 0 {
		commandArgs = env.BuildInteractiveWrapper(shell, "")
	}

	return &WorkflowResult{
		SessionName: sessionName,
		Path:        item.Path,
		CommandArgs: commandArgs,
	}, nil
}
