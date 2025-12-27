package cli

import (
	"atelier-go/internal/config"
	"atelier-go/internal/locations"
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newLocationsCmd() *cobra.Command {
	var listProjectsOnly bool
	var listZoxideOnly bool

	cmd := &cobra.Command{
		Use:   "locations",
		Short: "List available projects and directories",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load()
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to load config: %v\n", err)
			}

			projectProvider := locations.NewProjectProvider(cfg)
			zoxideProvider := locations.NewZoxideProvider()
			manager := locations.NewManager(projectProvider, zoxideProvider)

			locs, err := manager.GetAll(context.Background())
			if err != nil {
				fmt.Fprintf(os.Stderr, "error fetching locations: %v\n", err)
				os.Exit(1)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			if _, err := fmt.Fprintln(w, "SOURCE\tNAME\tPATH\tACTIONS"); err != nil {
				fmt.Fprintf(os.Stderr, "error writing to stdout: %v\n", err)
				return
			}

			for _, loc := range locs {
				if listProjectsOnly && loc.Source != "Project" {
					continue
				}
				if listZoxideOnly && loc.Source != "Zoxide" {
					continue
				}

				actionCount := len(loc.Actions)
				actionStr := "-"
				if actionCount > 0 {
					actionStr = fmt.Sprintf("%d", actionCount)
				}

				if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", loc.Source, loc.Name, loc.Path, actionStr); err != nil {
					fmt.Fprintf(os.Stderr, "error writing to stdout: %v\n", err)
					return
				}
			}
			if err := w.Flush(); err != nil {
				fmt.Fprintf(os.Stderr, "error flushing stdout: %v\n", err)
			}
		},
	}

	cmd.Flags().BoolVarP(&listProjectsOnly, "projects", "p", false, "List only configured projects")
	cmd.Flags().BoolVarP(&listZoxideOnly, "zoxide", "z", false, "List only zoxide directories")

	return cmd
}
