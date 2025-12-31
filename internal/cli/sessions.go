package cli

import (
	"atelier-go/internal/config"
	"atelier-go/internal/env"
	"atelier-go/internal/locations"
	"atelier-go/internal/sessions"
	"atelier-go/internal/utils"
	"context"
	"fmt"
	"os"
	"path/filepath"

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
	var projectFlag string
	var actionFlag string
	var folderFlag string

	cmd := &cobra.Command{
		Use:   "attach",
		Short: "Attach to a session",
		Long:  "Attach to a session using --project or --folder. If using --project, you can optionally specify an --action.",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.LoadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
				os.Exit(1)
			}

			target, err := resolveTarget(cmd.Context(), cfg, projectFlag, folderFlag, actionFlag)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}

			sessionManager := sessions.NewManager()
			if err := sessionManager.Attach(target.Name, target.Path, target.Command...); err != nil {
				fmt.Fprintf(os.Stderr, "error attaching to session: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVarP(&projectFlag, "project", "p", "", "Project name to attach to")
	cmd.Flags().StringVarP(&actionFlag, "action", "a", "", "Action name to run (optional, used with --project)")
	cmd.Flags().StringVarP(&folderFlag, "folder", "f", "", "Folder path to attach to")

	return cmd
}

func resolveTarget(ctx context.Context, cfg *config.Config, projectName, folderPath, actionName string) (*sessions.Target, error) {
	var loc *locations.Location

	if projectName != "" {
		locMgr, err := setupLocationManager(cfg, true, false)
		if err != nil {
			return nil, fmt.Errorf("failed to setup location manager: %w", err)
		}
		l, err := locMgr.Find(ctx, projectName)
		if err != nil {
			return nil, err
		}
		loc = l
	} else if folderPath != "" {
		absPath, err := utils.ExpandPath(folderPath)
		if err != nil {
			return nil, fmt.Errorf("failed to expand path: %w", err)
		}
		absPath, err = filepath.Abs(absPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path: %w", err)
		}

		loc = &locations.Location{
			Name:   filepath.Base(absPath),
			Path:   absPath,
			Source: "Folder",
		}
	} else {
		return nil, fmt.Errorf("must provide --project or --folder")
	}

	shell := env.DetectShell()
	return sessions.NewManager().Resolve(*loc, actionName, shell, cfg.GetEditor())
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
