package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"atelier-go/internal/engine"
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

		// 1. Try to resolve target from known items (Config/Zoxide)
		items, err := engine.Fetch()
		if err != nil {
			// Log warning but proceed to try as a raw path
			fmt.Fprintf(os.Stderr, "warning: could not fetch projects: %v\n", err)
		}

		var foundItem *engine.Item

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
			// If > 1 actions, we default to shell (name is already set to sanitized project name)
			// effectively behaving like "Default Shell" unless we add support for specifying action.

		} else {
			// 2. If not found, treat as a direct path
			absPath, err := filepath.Abs(target)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error resolving path '%s': %v\n", target, err)
				os.Exit(1)
			}
			path = absPath
			name = zmx.Sanitize(filepath.Base(absPath))
		}

		// 3. Attach
		fmt.Printf("Attaching to session '%s' in %s\n", name, path)
		if err := zmx.Attach(name, path, commandArgs...); err != nil {
			fmt.Fprintf(os.Stderr, "error attaching to session: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	attachCmd.Flags().BoolVarP(&attachShell, "shell", "s", false, "Force attach to shell session")
	rootCmd.AddCommand(attachCmd)
}
