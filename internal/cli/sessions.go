package cli

import (
	"atelier-go/internal/sessions"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newSessionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sessions",
		Short: "Manage workspace sessions",
	}

	cmd.AddCommand(newSessionsAttachCmd())
	cmd.AddCommand(newSessionsKillCmd())
	cmd.AddCommand(newSessionsListCmd())

	return cmd
}

func newSessionsAttachCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "attach [name] [path]",
		Short: "Attach to a session",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			path := args[1]

			// Optional command args
			var cmdArgs []string
			if len(args) > 2 {
				cmdArgs = args[2:]
			}

			manager := sessions.NewManager()
			if err := manager.Attach(name, path, cmdArgs...); err != nil {
				fmt.Fprintf(os.Stderr, "error attaching to session: %v\n", err)
				os.Exit(1)
			}
		},
	}
}

func newSessionsKillCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "kill [name]",
		Short: "Kill a session",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			manager := sessions.NewManager()
			if err := manager.Kill(name); err != nil {
				fmt.Fprintf(os.Stderr, "error killing session: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Session '%s' killed.\n", name)
		},
	}
}

func newSessionsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List active sessions",
		Run: func(cmd *cobra.Command, args []string) {
			manager := sessions.NewManager()
			sessList, err := manager.List()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error listing sessions: %v\n", err)
				os.Exit(1)
			}

			if err := manager.PrintTable(os.Stdout, sessList); err != nil {
				fmt.Fprintf(os.Stderr, "error printing sessions: %v\n", err)
			}
		},
	}
}
