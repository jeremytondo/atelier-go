package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"atelier-go/internal/locations"
	"atelier-go/internal/ui"
	"atelier-go/internal/zmx"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "atelier-go",
	Short: "A local-first CLI workflow tool",
	Run:   run,
}

func main() {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	items, err := locations.Fetch()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error fetching items: %v\n", err)
		os.Exit(1)
	}

	if len(items) == 0 {
		fmt.Println("No projects or recent directories found.")
		return
	}

	displayMap := make(map[string]locations.Item)
	var choices []string

	for _, item := range items {
		label := ui.FormatItem(item)
		choices = append(choices, label)
		// Use stripped label as key for robustness
		displayMap[ui.StripANSI(label)] = item
	}

	selection, key, err := ui.Select(choices, "Select Project (Alt-s for Shell)", "Project ➜ ", []string{"alt-s"})
	if err != nil {
		return
	}

	item, ok := displayMap[ui.StripANSI(selection)]
	if !ok {
		// Fallback: Try exact match
		if i, okExact := displayMap[selection]; okExact {
			item = i
		} else {
			fmt.Fprintf(os.Stderr, "error: invalid selection: %q\n", selection)
			os.Exit(1)
		}
	}

	sessionName := zmx.Sanitize(item.Name)
	var commandArgs []string
	forceShell := key == "alt-s"

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

	manager := zmx.New()
	fmt.Printf("Attaching to session '%s' in %s\n", sessionName, item.Path)
	if err := manager.Attach(sessionName, item.Path, commandArgs...); err != nil {
		fmt.Fprintf(os.Stderr, "error attaching to session: %v\n", err)
		os.Exit(1)
	}
}
