package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"text/template"
	"time"

	"atelier-go/internal/api"
	"atelier-go/internal/auth"
	"atelier-go/internal/system"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func printStartupBanner(addr, token string) {
	fmt.Println("┌──────────────────────────────────────────────────────────────┐")
	fmt.Println("│  Atelier Server started!                                     │")
	fmt.Printf("│  Address: %-51s│\n", addr)
	fmt.Println("│                                                              │")
	fmt.Println("│  Token:                                                      │")
	fmt.Printf("│  %s                            │\n", token)
	fmt.Println("│                                                              │")
	fmt.Println("│  To connect, run on client:                                  │")
	fmt.Printf("│  atelier-go client login %s    │\n", token)
	fmt.Println("└──────────────────────────────────────────────────────────────┘")
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Atelier daemon",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		system.LoadConfig("server")
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Setup Logger
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		slog.SetDefault(logger)

		host := viper.GetString("host")
		if host == "" {
			host = "0.0.0.0"
		}
		port := viper.GetInt("port")
		addr := fmt.Sprintf("%s:%d", host, port)

		// Initialize Token
		tokenPath, err := auth.GetDefaultTokenPath()
		if err != nil {
			logger.Error("Failed to get token path", "error", err)
			os.Exit(1)
		}
		token, err := auth.LoadOrCreateToken(tokenPath)
		if err != nil {
			logger.Error("Failed to initialize token", "error", err)
			os.Exit(1)
		}
		authenticator := auth.NewAuthenticator(token)

		printStartupBanner(addr, token)

		// Load Configuration
		var actions []api.Action
		if err := viper.UnmarshalKey("actions", &actions); err != nil {
			logger.Warn("Failed to load actions config", "error", err)
			actions = []api.Action{}
		}

		// Create Server
		serverConfig := api.Config{
			Actions: actions,
		}
		apiServer := api.NewServer(serverConfig)

		// Setup Router
		handler := apiServer.Routes(authenticator.Middleware)

		server := &http.Server{
			Addr:    addr,
			Handler: handler,
		}

		// Start Server in Goroutine
		go func() {
			logger.Info("Starting server", "address", addr)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("Server failed", "error", err)
				os.Exit(1)
			}
		}()

		// Wait for Interrupt Signal
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

		<-stop
		logger.Info("Shutting down server...")

		// Context with timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Server forced to shutdown", "error", err)
		}

		// Cleanup PID file if it exists (though strictly speaking, the start command manages the PID file,
		// but if we are running in foreground or if start command exec'd us, we might want to ensure cleanup)
		// The `start` command writes the PID file.
		if err := system.RemovePIDFile(); err != nil {
			// It's okay if it doesn't exist
			if !os.IsNotExist(err) {
				logger.Warn("Failed to remove PID file", "error", err)
			}
		}

		logger.Info("Server exited")
	},
}

var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server in the background",
	Run: func(cmd *cobra.Command, args []string) {
		// Check if already running
		pid, err := system.ReadPIDFile()
		if err == nil {
			if system.IsProcessRunning(pid) {
				fmt.Printf("Server is already running (PID: %d)\n", pid)
				return
			}
			// PID file exists but process is dead, clean it up
			system.RemovePIDFile()
		}

		// Find the executable
		exe, err := os.Executable()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to determine executable path: %v\n", err)
			os.Exit(1)
		}

		// Start the process detached
		startCmd := exec.Command(exe, "server")
		// Detach logic usually involves redirecting std I/O and handling process groups,
		// but simply starting it via exec without waiting and letting it write its PID is often enough for simple use cases.
		// However, to truly background it and have it survive the parent shell exit, we might want to setsid.
		// For simplicity in this Go implementation, we'll start it and just not Wait().
		// Since the parent (this CLI command) exits immediately, the child effectively becomes a daemon.

		// Redirect output to log file? Or just /dev/null for now?
		// A proper background service should probably log somewhere.
		// For now, let's inherit environment but detach I/O so it doesn't hang the terminal.
		startCmd.Stdout = nil
		startCmd.Stderr = nil
		startCmd.Stdin = nil

		if err := startCmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
			os.Exit(1)
		}

		// Write PID file
		if err := system.WritePIDFile(startCmd.Process.Pid); err != nil {
			// Try to kill the process if we can't write the PID file
			startCmd.Process.Kill()
			fmt.Fprintf(os.Stderr, "Failed to write PID file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Server started in background (PID: %d)\n", startCmd.Process.Pid)

		// Print the banner for the user
		host := viper.GetString("host")
		if host == "" {
			host = "0.0.0.0"
		}
		port := viper.GetInt("port")
		addr := fmt.Sprintf("%s:%d", host, port)

		tokenPath, err := auth.GetDefaultTokenPath()
		if err == nil {
			token, err := auth.LoadOrCreateToken(tokenPath)
			if err == nil {
				printStartupBanner(addr, token)
			}
		}
	},
}

var serverStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the background server",
	Run: func(cmd *cobra.Command, args []string) {
		pid, err := system.ReadPIDFile()
		if err != nil {
			fmt.Println("No running server found (PID file missing).")
			return
		}

		proc, err := os.FindProcess(pid)
		if err != nil {
			fmt.Println("Could not find process.")
			// Clean up stale file
			system.RemovePIDFile()
			return
		}

		if err := proc.Kill(); err != nil {
			fmt.Printf("Failed to stop server: %v\n", err)
			return
		}

		system.RemovePIDFile()
		fmt.Println("Server stopped.")
	},
}

var serverInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the server as a systemd user service",
	Run: func(cmd *cobra.Command, args []string) {
		exe, err := os.Executable()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to determine executable path: %v\n", err)
			os.Exit(1)
		}

		// Template for systemd service
		const serviceTmpl = `[Unit]
Description=Atelier Go Server
Documentation=https://github.com/shell-pool/atelier-go
After=network.target

[Service]
Type=simple
ExecStart={{.ExecPath}} server
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=default.target
`
		data := struct {
			ExecPath string
		}{
			ExecPath: exe,
		}

		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get user home directory: %v\n", err)
			os.Exit(1)
		}

		// Ensure ~/.local/share/systemd/user exists
		serviceDir := filepath.Join(home, ".local", "share", "systemd", "user")
		if err := os.MkdirAll(serviceDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create systemd directory: %v\n", err)
			os.Exit(1)
		}

		servicePath := filepath.Join(serviceDir, "atelier-go.service")

		// Generate file
		f, err := os.Create(servicePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create service file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()

		tmpl, err := template.New("service").Parse(serviceTmpl)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse template: %v\n", err)
			os.Exit(1)
		}

		if err := tmpl.Execute(f, data); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write service file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Service file created at %s\n", servicePath)

		// Enable and start
		// systemctl --user daemon-reload
		// systemctl --user enable --now atelier-go

		if err := exec.Command("systemctl", "--user", "daemon-reload").Run(); err != nil {
			fmt.Printf("Warning: Failed to reload systemd daemon: %v\n", err)
		} else {
			if err := exec.Command("systemctl", "--user", "enable", "--now", "atelier-go").Run(); err != nil {
				fmt.Printf("Warning: Failed to enable/start service: %v\n", err)
			} else {
				fmt.Println("Service installed and started successfully!")
			}
		}
	},
}

var serverTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Print the current authentication token",
	Run: func(cmd *cobra.Command, args []string) {
		tokenPath, err := auth.GetDefaultTokenPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get token path: %v\n", err)
			os.Exit(1)
		}

		token, err := auth.LoadOrCreateToken(tokenPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load token: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(token)
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

		tokenPath, err := auth.GetDefaultTokenPath()
		var token string
		if err == nil {
			token, err = auth.LoadOrCreateToken(tokenPath)
		}
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
	serverCmd.AddCommand(serverTokenCmd)
	serverCmd.AddCommand(serverStartCmd)
	serverCmd.AddCommand(serverStopCmd)
	serverCmd.AddCommand(serverInstallCmd)
}
