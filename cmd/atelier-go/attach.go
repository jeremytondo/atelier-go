package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"atelier-go/internal/locations"
	"atelier-go/internal/zmx"
)

var attachShell bool

var attachCmd = &cobra.Command{
	Use:   "attach [target]",
	Short: "Attach to a project or directory session",
	Long:  `Attach to a project by name or path. If the target matches a known project or Zoxide entry, it uses that. Otherwise, it treats the argument as a path.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := args[0]

		items, err := locations.Fetch()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not fetch projects: %v\n", err)
		}

		var foundItem *locations.Item

		for i := range items {
			if items[i].Name == target || items[i].Path == target {
				foundItem = &items[i]
				break
			}
		}

		var name, path string
		var commandArgs []string

		if foundItem != nil {
			name = zmx.Sanitize(foundItem.Name)
			path = foundItem.Path

			// Logic for 1 action auto-attach (unless shell is forced)
			if !attachShell && len(foundItem.Actions) == 1 {
				act := foundItem.Actions[0]
				name = fmt.Sprintf("%s:%s", name, zmx.Sanitize(act.Name))
				commandArgs = strings.Fields(act.Command)
			}
		} else {
			// Resolve as direct path
			absPath, err := filepath.Abs(target)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error resolving path '%s': %v\n", target, err)
				os.Exit(1)
			}
			path = absPath
			name = zmx.Sanitize(filepath.Base(absPath))
		}

		manager := zmx.New()
		fmt.Printf("Attaching to session '%s' in %s\n", name, path)
		if err := manager.Attach(name, path, commandArgs...); err != nil {
			fmt.Fprintf(os.Stderr, "error attaching to session: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	attachCmd.Flags().BoolVarP(&attachShell, "shell", "s", false, "Force attach to shell session")
	rootCmd.AddCommand(attachCmd)
}
