package cli

import (
	"atelier-go/internal/config"
	"atelier-go/internal/ui"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newUICmd() *cobra.Command {
	var showProjects bool
	var showZoxide bool

	cmd := &cobra.Command{
		Use:     "ui",
		Aliases: []string{"start"},
		Short:   "Start the interactive UI with custom options",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.LoadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
				os.Exit(1)
			}

			clientID, _ := cmd.Flags().GetString("client-id")

			mgr, err := setupLocationManager(cfg, showProjects, showZoxide)
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

	cmd.Flags().BoolVarP(&showProjects, "projects", "p", false, "Show projects only")
	cmd.Flags().BoolVarP(&showZoxide, "zoxide", "z", false, "Show zoxide directories only")

	return cmd
}
