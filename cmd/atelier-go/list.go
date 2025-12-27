package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"atelier-go/internal/locations"
)

var (
	listProjectsOnly bool
	listZoxideOnly   bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available projects and directories",
	Run: func(cmd *cobra.Command, args []string) {
		items, err := locations.Fetch()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error fetching items: %v\n", err)
			os.Exit(1)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		if _, err := fmt.Fprintln(w, "SOURCE\tNAME\tPATH\tACTIONS"); err != nil {
			fmt.Fprintf(os.Stderr, "error writing to stdout: %v\n", err)
			return
		}

		for _, item := range items {
			if listProjectsOnly && item.Source != "Project" {
				continue
			}
			if listZoxideOnly && item.Source != "Zoxide" {
				continue
			}

			actionCount := len(item.Actions)
			actionStr := "-"
			if actionCount > 0 {
				actionStr = fmt.Sprintf("%d", actionCount)
			}

			if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", item.Source, item.Name, item.Path, actionStr); err != nil {
				fmt.Fprintf(os.Stderr, "error writing to stdout: %v\n", err)
				return
			}
		}
		if err := w.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "error flushing writer: %v\n", err)
		}
	},
}

func init() {
	listCmd.Flags().BoolVarP(&listProjectsOnly, "projects", "p", false, "List only configured projects")
	listCmd.Flags().BoolVarP(&listZoxideOnly, "zoxide", "z", false, "List only zoxide directories")
	rootCmd.AddCommand(listCmd)
}
