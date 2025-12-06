package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Connect to the Atelier daemon",
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
		if host == "" {
			host = "localhost"
		}
		port := viper.GetInt("port")

		fmt.Printf("Client Configuration:\nHost: %s\nPort: %d\n", host, port)
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)
}
