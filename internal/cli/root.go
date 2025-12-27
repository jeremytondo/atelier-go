package cli

import (
	"context"
	"fmt"
	"os"

	"atelier-go/internal/config"
	"atelier-go/internal/locations"
	"atelier-go/internal/sessions"
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
		Run:     runRoot,
	}

	cmd.AddCommand(newLocationsCmd())
	cmd.AddCommand(newSessionsCmd())

	return cmd
}

func runRoot(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to load config: %v\n", err)
	}

	// Initialize providers
	projectProvider := locations.NewProjectProvider(cfg)
	zoxideProvider := locations.NewZoxideProvider()
	manager := locations.NewManager(projectProvider, zoxideProvider)

	// Fetch locations
	locs, err := manager.GetAll(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error fetching locations: %v\n", err)
		os.Exit(1)
	}

	if len(locs) == 0 {
		fmt.Println("No projects or recent directories found.")
		return
	}

	// Run interactive workflow
	result, err := ui.RunInteractiveWorkflow(locs)
	if err != nil {
		// If cancelled (e.g. exit code 1 or 130), just exit quietly or print error if meaningful
		if result == nil {
			return
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}

	// Attach to session
	sessionManager := sessions.NewManager()
	fmt.Printf("Attaching to session '%s' in %s\n", result.SessionName, result.Path)
	if err := sessionManager.Attach(result.SessionName, result.Path, result.CommandArgs...); err != nil {
		fmt.Fprintf(os.Stderr, "error attaching to session: %v\n", err)
		os.Exit(1)
	}
}
