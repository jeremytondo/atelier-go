package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Atelier daemon",
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
		if host == "" {
			host = "0.0.0.0"
		}
		port := viper.GetInt("port")

		fmt.Printf("Server Configuration:\nHost: %s\nPort: %d\n", host, port)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
