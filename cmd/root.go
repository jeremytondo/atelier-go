package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "atelier-go",
	Short: "Atelier Go - Remote Development Daemon & Client",
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
