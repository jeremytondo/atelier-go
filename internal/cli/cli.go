// Package cli implements the Cobra commands for the application.
package cli

import (
	"context"
	"fmt"
	"os"

	"atelier-go/internal/ui"

	"github.com/spf13/cobra"
)

var version = "dev"

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(ctx context.Context) {
	if err := newRootCmd().ExecuteContext(ctx); err != nil {
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
			// Default behavior: UI handles defaults (showing everything)
			if err := ui.Run(cmd.Context(), ui.Options{}); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.AddCommand(newUICmd())
	cmd.AddCommand(newLocationsCmd())
	cmd.AddCommand(newSessionsCmd())
	cmd.AddCommand(newHostnameCmd())

	return cmd
}
