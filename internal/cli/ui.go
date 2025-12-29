package cli

import (
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
			mgr, err := setupLocationManager(showProjects, showZoxide)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			if err := ui.Run(cmd.Context(), mgr); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().BoolVarP(&showProjects, "projects", "p", false, "Show projects only")
	cmd.Flags().BoolVarP(&showZoxide, "zoxide", "z", false, "Show zoxide directories only")

	return cmd
}
