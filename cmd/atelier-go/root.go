package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"atelier-go/internal/engine"
	"atelier-go/internal/ui"
	"atelier-go/internal/zmx"
)

var rootCmd = &cobra.Command{
	Use:   "atelier-go",
	Short: "A local-first CLI workflow tool",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	// 1. Fetch Items
	items, err := engine.Fetch()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error fetching items: %v\n", err)
		os.Exit(1)
	}

	if len(items) == 0 {
		fmt.Println("No projects or recent directories found.")
		return
	}

	// 2. Format for UI
	// We map the display string back to the item to avoid parsing logic
	displayMap := make(map[string]engine.Item)
	var choices []string

	const (
		projectColor = "\x1b[38;2;137;180;250m"
		resetColor   = "\x1b[0m"
	)

	for _, item := range items {
		// align columns if possible, but for now simple format
		// [Source] Name (Path)
		cleanLabel := fmt.Sprintf("[%s] %s  (%s)", item.Source, item.Name, item.Path)
		displayLabel := cleanLabel

		if item.Source == "Project" {
			displayLabel = fmt.Sprintf("%s%s%s", projectColor, cleanLabel, resetColor)
		}

		choices = append(choices, displayLabel)
		displayMap[cleanLabel] = item
	}

	// 3. Select Project
	selection, key, err := ui.Select(choices, "Select Project (Alt-s for Shell)", "Project ➜ ", []string{"alt-s"})
	if err != nil {
		// If cancelled or empty, just exit quietly
		return
	}

	item, ok := displayMap[selection]
	if !ok {
		fmt.Fprintf(os.Stderr, "error: invalid selection\n")
		os.Exit(1)
	}

	sessionName := zmx.Sanitize(item.Name)
	var commandArgs []string
	forceShell := key == "alt-s"

	// 4. Select Action (if applicable)
	if !forceShell && len(item.Actions) > 0 {
		if len(item.Actions) == 1 {
			// Auto-attach to the single configured action
			act := item.Actions[0]
			sessionName = fmt.Sprintf("%s:%s", zmx.Sanitize(item.Name), zmx.Sanitize(act.Name))
			commandArgs = strings.Fields(act.Command)
		} else {
			// Multiple actions: Show menu
			type actEntry struct {
				cmd  string
				name string // sanitized action suffix
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
				actionMap[label] = actEntry{cmd: act.Command, name: zmx.Sanitize(act.Name)}
			}

			actSelection, _, err := ui.Select(actionChoices, fmt.Sprintf("Select Action for %s", item.Name), "Action ➜ ", nil)
			if err != nil {
				return
			}

			if entry, ok := actionMap[actSelection]; ok {
				if entry.name != "" {
					sessionName = fmt.Sprintf("%s:%s", zmx.Sanitize(item.Name), entry.name)
				}
				if entry.cmd != "" {
					commandArgs = strings.Fields(entry.cmd)
				}
			}
		}
	}

	// 5. Attach
	fmt.Printf("Attaching to session '%s' in %s\n", sessionName, item.Path)
	if err := zmx.Attach(sessionName, item.Path, commandArgs...); err != nil {
		fmt.Fprintf(os.Stderr, "error attaching to session: %v\n", err)
		os.Exit(1)
	}
}
