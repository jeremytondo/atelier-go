package ui

import (
	"atelier-go/internal/locations"
	"atelier-go/internal/sessions"
	"fmt"
	"strings"
)

// WorkflowResult represents the outcome of the user interaction.
type WorkflowResult struct {
	SessionName string
	Path        string
	CommandArgs []string
}

// RunInteractiveWorkflow executes the interactive selection process.
func RunInteractiveWorkflow(locs []locations.Location) (*WorkflowResult, error) {
	if len(locs) == 0 {
		return nil, fmt.Errorf("no locations found")
	}

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
					commandArgs = strings.Fields(entry.cmd)
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
