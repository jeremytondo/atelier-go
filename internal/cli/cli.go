// Package cli implements the Cobra commands for the application.
package cli

import (
	"fmt"
	"os"

	"atelier-go/internal/ui"

	"github.com/spf13/cobra"
)

var version = "dev"

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "atelier-go",
		Short:   "A local-first CLI workflow tool",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			// Default to showing everything
			opts := ui.Options{
				ShowProjects: true,
				ShowZoxide:   true,
			}
			if err := ui.Run(opts); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.AddCommand(newUiCmd())
	cmd.AddCommand(newLocationsCmd())
	cmd.AddCommand(newSessionsCmd())

	return cmd
}
