package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "atelier-go",
	Short: "Atelier Go - Remote Development Server & Client",
	Long: `Atelier Go is a unified tool for managing remote development environments.

It consists of two parts:
1. A Server that runs on your development machine/host.
2. A Client that you run to connect, manage, and attach to sessions.

It integrates with 'shpool' for persistent sessions and 'zoxide' for smart path navigation.`,
	Example: `  # Start the server in the background
  atelier-go server start

  # Connect to the server to select a session or project
  atelier-go client

  # Check server status
  atelier-go server status`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags (optional, but good practice to map them)
	rootCmd.PersistentFlags().String("host", "", "Host to connect to (client) or bind to (server)")
	rootCmd.PersistentFlags().Int("port", 9001, "Port to connect to or listen on")

	// Bind flags to viper
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
}

func initConfig() {
	// Set environment prefix
	viper.SetEnvPrefix("ATELIER")

	// Check for env vars (e.g. ATELIER_HOST, ATELIER_PORT)
	viper.AutomaticEnv()
}
