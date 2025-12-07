package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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

var serverStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the status of the Atelier daemon",
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
		if host == "" || host == "0.0.0.0" {
			host = "127.0.0.1"
		}
		port := viper.GetInt("port")
		url := fmt.Sprintf("http://%s:%d/health", host, port)

		tokenPath := os.ExpandEnv("$HOME/.config/atelier/token")
		token, err := auth.LoadOrCreateToken(tokenPath)
		if err != nil {
			fmt.Printf("Warning: Failed to load token: %v\n", err)
		}

		client := &http.Client{
			Timeout: 2 * time.Second,
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			return
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Server is DOWN")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			fmt.Println("Server is UP")
			fmt.Printf("Address: http://%s:%d\n", host, port)
		} else {
			fmt.Printf("Server is UP but returned status: %d\n", resp.StatusCode)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.AddCommand(serverStatusCmd)
}
