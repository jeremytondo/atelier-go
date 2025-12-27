package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"atelier-go/internal/engine"
	"atelier-go/internal/zmx"
)

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

		var name, path string
		var found bool

		for _, item := range items {
			if item.Name == target || item.Path == target {
				name = item.Name
				path = item.Path
				found = true
				break
			}
		}

		// 2. If not found, treat as a direct path
		if !found {
			absPath, err := filepath.Abs(target)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error resolving path '%s': %v\n", target, err)
				os.Exit(1)
			}
			path = absPath
			name = filepath.Base(absPath)
		}

		// 3. Attach
		fmt.Printf("Attaching to session '%s' in %s\n", name, path)
		if err := zmx.Attach(name, path); err != nil {
			fmt.Fprintf(os.Stderr, "error attaching to session: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(attachCmd)
}
