package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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

func loadConfig(configName string) {
	// 1. Determine the config directory
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			// If we can't find home, we can't load config.
			// Just return and rely on defaults/flags.
			return
		}
		configDir = filepath.Join(home, ".config")
	}

	atelierConfigDir := filepath.Join(configDir, "atelier-go")

	// 2. Setup Viper
	viper.AddConfigPath(atelierConfigDir)
	viper.SetConfigName(configName)
	viper.SetConfigType("toml")

	// 3. Read Config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but had an error (e.g. invalid syntax)
			fmt.Fprintf(os.Stderr, "Warning: failed to read config file: %v\n", err)
		}
	}
}
