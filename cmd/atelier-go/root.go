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

	for _, item := range items {
		// align columns if possible, but for now simple format
		// [Source] Name (Path)
		label := fmt.Sprintf("[%s] %s  (%s)", item.Source, item.Name, item.Path)
		choices = append(choices, label)
		displayMap[label] = item
	}

	// 3. Select
	selection, err := ui.Select(choices, "Select Project", "Project ➜ ")
	if err != nil {
		// If cancelled or empty, just exit quietly
		return
	}

	item, ok := displayMap[selection]
	if !ok {
		fmt.Fprintf(os.Stderr, "error: invalid selection\n")
		os.Exit(1)
	}

	// 4. Select Action
	var commandArgs []string

	if len(item.Actions) > 0 {
		actionMap := make(map[string]string)
		var actionChoices []string

		// Add default Shell option
		shellLabel := "Shell (Default)"
		actionChoices = append(actionChoices, shellLabel)
		actionMap[shellLabel] = ""

		for _, act := range item.Actions {
			label := fmt.Sprintf("%s (%s)", act.Name, act.Command)
			actionChoices = append(actionChoices, label)
			actionMap[label] = act.Command
		}

		actSelection, err := ui.Select(actionChoices, fmt.Sprintf("Select Action for %s", item.Name), "Action ➜ ")
		if err != nil {
			return
		}

		if cmdStr, ok := actionMap[actSelection]; ok && cmdStr != "" {
			commandArgs = strings.Fields(cmdStr)
		}
	}

	// 5. Attach
	if err := zmx.Attach(item.Name, item.Path, commandArgs...); err != nil {
		fmt.Fprintf(os.Stderr, "error attaching to session: %v\n", err)
		os.Exit(1)
	}
}
