package cli

import (
	"atelier-go/internal/config"
	"atelier-go/internal/locations"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newLocationsCmd() *cobra.Command {
	var listProjectsOnly bool
	var listZoxideOnly bool

	cmd := &cobra.Command{
		Use:   "locations",
		Short: "List available projects and directories",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.LoadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
				os.Exit(1)
			}

			mgr, err := setupLocationManager(cfg, listProjectsOnly, listZoxideOnly)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error building manager: %v\n", err)
				os.Exit(1)
			}

			locs, err := mgr.GetAll(cmd.Context())
			if err != nil {
				fmt.Fprintf(os.Stderr, "error fetching locations: %v\n", err)
				os.Exit(1)
			}

			if err := locations.PrintTable(os.Stdout, locs); err != nil {
				fmt.Fprintf(os.Stderr, "error printing locations: %v\n", err)
			}
		},
	}

	cmd.Flags().BoolVarP(&listProjectsOnly, "projects", "p", false, "List only configured projects")
	cmd.Flags().BoolVarP(&listZoxideOnly, "zoxide", "z", false, "List only zoxide directories")

	return cmd
}
