package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"atelier-go/internal/api"
	"atelier-go/internal/auth"

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
		addr := fmt.Sprintf("%s:%d", host, port)

		// Initialize Token
		tokenPath := os.ExpandEnv("$HOME/.config/atelier/token")
		token, err := auth.LoadOrCreateToken(tokenPath)
		if err != nil {
			log.Fatalf("Failed to initialize token: %v", err)
		}

		fmt.Printf("Atelier Server starting on %s\n", addr)
		fmt.Printf("Token: %s\n", token)

		// Setup Router
		mux := http.NewServeMux()

		// Health Endpoint (Protected)
		mux.Handle("/health", auth.RequireToken(http.HandlerFunc(api.HealthHandler)))

		// Locations Endpoint (Protected)
		mux.Handle("/api/locations", auth.RequireToken(http.HandlerFunc(api.LocationsHandler)))

		server := &http.Server{
			Addr:    addr,
			Handler: mux,
		}

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
