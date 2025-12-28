// Package ui provides the interactive user interface.
package ui

import (
	"atelier-go/internal/locations"
	"atelier-go/internal/sessions"
	"context"
	"fmt"
	"strings"
)

// Options configuration for the UI.
type Options struct {
	ShowProjects bool
	ShowZoxide   bool
}

// Run executes the interactive UI.
// It fetches locations based on options, prompts the user, and attaches to a session.
func Run(ctx context.Context, opts Options) error {
	// 1. Fetch Locations
	fetchOpts := locations.FetchOptions{
		IncludeProjects: opts.ShowProjects,
		IncludeZoxide:   opts.ShowZoxide,
	}

	locs, err := locations.List(ctx, fetchOpts)
	if err != nil {
		return fmt.Errorf("failed to fetch locations: %w", err)
	}

	if len(locs) == 0 {
		fmt.Println("No projects or recent directories found.")
		return nil
	}

	// 2. Interactive Selection
	result, err := runSelection(locs)
	if err != nil {
		return err
	}
	if result == nil {
		// User cancelled
		return nil
	}

	// 3. Attach to Session
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
func runSelection(locs []locations.Location) (*WorkflowResult, error) {
	displayMap := make(map[string]locations.Location)
	var choices []string

	for _, loc := range locs {
		label := FormatLocation(loc)
		choices = append(choices, label)
		// Use stripped label as key for robustness
		displayMap[StripANSI(label)] = loc
	}

	selection, key, err := Select(choices, "Select Project (Alt-s for Shell)", "Project ➜ ", []string{"alt-s"})
	if err != nil {
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

	sessionName := sessions.Sanitize(item.Name)
	var commandArgs []string
	forceShell := key == "alt-s"

	if !forceShell && len(item.Actions) > 0 {
		if len(item.Actions) == 1 {
			// Auto-attach to the single configured action
			act := item.Actions[0]
			sessionName = fmt.Sprintf("%s:%s", sessions.Sanitize(item.Name), sessions.Sanitize(act.Name))
			commandArgs = strings.Fields(act.Command)
		} else {
			// Multiple actions: Show menu
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
				label := fmt.Sprintf("%s (%s)", act.Name, act.Command)
				actionChoices = append(actionChoices, label)
				actionMap[label] = actEntry{cmd: act.Command, name: sessions.Sanitize(act.Name)}
			}

			actSelection, _, err := Select(actionChoices, fmt.Sprintf("Select Action for %s", item.Name), "Action ➜ ", nil)
			if err != nil {
				return nil, err
			}

			if entry, ok := actionMap[actSelection]; ok {
				if entry.name != "" {
					sessionName = fmt.Sprintf("%s:%s", sessions.Sanitize(item.Name), entry.name)
				}
				if entry.cmd != "" {
					// Use shell to execute command string
					commandArgs = []string{"/bin/sh", "-c", entry.cmd}
				}
			}
		}
	}

	return &WorkflowResult{
		SessionName: sessionName,
		Path:        item.Path,
		CommandArgs: commandArgs,
	}, nil
}
