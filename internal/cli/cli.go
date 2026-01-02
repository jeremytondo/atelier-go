// Package cli implements the Cobra commands for the application.
package cli

import (
	"context"
	"fmt"
	"os"

	"atelier-go/internal/config"
	"atelier-go/internal/ui"

	"github.com/spf13/cobra"
)

var Version = "dev"

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
		Version: Version,
		Run: func(cmd *cobra.Command, args []string) {
			// Default behavior: UI handles defaults (showing everything)
			cfg, err := config.LoadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
				os.Exit(1)
			}

			clientID, _ := cmd.Flags().GetString("client-id")

			mgr, err := setupLocationManager(cfg, false, false)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			if err := ui.Run(cmd.Context(), mgr, cfg, clientID); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.PersistentFlags().String("client-id", "", "Client identifier for session recovery")

	cmd.AddCommand(newUICmd())
	cmd.AddCommand(newLocationsCmd())
	cmd.AddCommand(newSessionsCmd())
	cmd.AddCommand(newHostnameCmd())

	return cmd
}
