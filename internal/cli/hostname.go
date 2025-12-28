package cli

import (
	"atelier-go/internal/utils"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newHostnameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hostname",
		Short: "Print the current machine hostname used for configuration",
		Run: func(cmd *cobra.Command, args []string) {
			hostname, err := utils.GetHostname()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error determining hostname: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(hostname)
		},
	}
	return cmd
}
