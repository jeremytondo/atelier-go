package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"atelier-go/internal/api"
	"atelier-go/internal/auth"
	"atelier-go/internal/client"
	"atelier-go/internal/system"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	iconSession = " "
	iconProject = "\ueb30 "
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Connect to the server and manage sessions",
	Long: `The client command connects to the running Atelier server.

It presents an interactive list of:
- Active 'shpool' sessions
- Defined projects
- Frequent directories (via zoxide)

You can filter this list using flags. Selecting an item will either attach
to an existing session or start a new one in that location.`,
	Example: `  # Default: Show sessions, projects, and frequent paths
  atelier-go client

  # Show only active sessions
  atelier-go client --sessions

  # Show ALL directories (can be slow)
  atelier-go client --all`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		system.LoadConfig("client")
	},
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")
		sessions, _ := cmd.Flags().GetBool("sessions")

		filter := "frequent"
		if sessions {
			filter = "sessions"
		} else if all {
			filter = "all"
		}
		runClient(filter)
	},
}

var clientLoginCmd = &cobra.Command{
	Use:   "login [token]",
	Short: "Save the authentication token",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		token := args[0]
		tokenPath, err := auth.GetDefaultTokenPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting token path: %v\n", err)
			os.Exit(1)
		}

		if err := auth.SaveToken(tokenPath, token); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving token: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Token saved successfully to %s\n", tokenPath)
	},
}

func init() {
	clientCmd.Flags().BoolP("all", "a", false, "Show all directories in home")
	clientCmd.Flags().BoolP("sessions", "s", false, "Show open sessions only")
	rootCmd.AddCommand(clientCmd)
	clientCmd.AddCommand(clientLoginCmd)
}

func runClient(filter string) {
	host := viper.GetString("host")
	if host == "" {
		host = "localhost"
	}
	port := viper.GetInt("port")

	// 1. Load Token
	tokenPath, err := auth.GetDefaultTokenPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting token path: %v\n", err)
		os.Exit(1)
	}
	token, err := auth.LoadOrCreateToken(tokenPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading token: %v\n", err)
		os.Exit(1)
	}

	// Initialize Client
	c := client.New(host, port, token)

	// 2. Fetch Locations
	locations, err := c.FetchLocations(filter)
	if err != nil {
		url := fmt.Sprintf("http://%s:%d", host, port)
		fmt.Fprintf(os.Stderr, "Error connecting to Atelier Daemon at %s: %v\n", url, err)
		fmt.Fprintf(os.Stderr, "Is the server running? (atelier-go server)\n")
		os.Exit(1)
	}

	// 3. Prepare List for FZF
	var options []string
	// Map to store project name -> location for lookup
	projectPaths := make(map[string]string)

	// Add Sessions first with icon
	for _, s := range locations.Sessions {
		options = append(options, iconSession+" "+s)
	}
	// Add Projects
	for _, p := range locations.Projects {
		// Use project name for display
		dispName := p.Name
		if dispName == "" {
			dispName = filepath.Base(p.Location)
		}

		// If duplicate names exist, we might overwrite, but that's a user config issue mostly.
		// To be safe we could append path if collision, but keeping it simple as requested.
		projectPaths[dispName] = p.Location

		options = append(options, iconProject+dispName)
	}
	// Add Paths
	options = append(options, locations.Paths...)

	if len(options) == 0 {
		fmt.Println("No locations or active sessions found.")
		os.Exit(0)
	}

	// 4. Select Location via FZF
	selection, err := runFzf(options, "Location ➜ ")
	if err != nil {
		// User likely cancelled
		os.Exit(0)
	}

	// 5. Handle Selection
	if sessionName, ok := strings.CutPrefix(selection, iconSession); ok {
		// Attach to existing session
		if err := c.Attach(strings.TrimSpace(sessionName)); err != nil {
			fmt.Fprintf(os.Stderr, "Error attaching to session: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Create new session at path
		path := selection
		if projName, ok := strings.CutPrefix(selection, iconProject); ok {
			// Resolve project name to path
			name := strings.TrimSpace(projName)
			if loc, found := projectPaths[name]; found {
				path = loc
			} else {
				// Fallback if something weird happens
				path = name
			}
		}

		// Fetch available actions from server
		actionsResp, err := c.FetchActions(path)
		var actions []api.Action
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to fetch actions: %v\n", err)
			// Fallback to default
			actions = []api.Action{
				{Name: "shell", Command: "$SHELL -l"},
			}
		} else {
			actions = actionsResp.Actions
		}

		action, err := selectAction(path, actions)
		if err != nil {
			os.Exit(0)
		}

		// Determine the name to use (project name or fallback to path base)
		name := filepath.Base(path)
		if strings.HasPrefix(selection, iconProject) {
			if projName, ok := strings.CutPrefix(selection, iconProject); ok {
				name = strings.TrimSpace(projName)
			}
		}

		if err := c.Start(path, name, action, actionsResp != nil && actionsResp.IsProject); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting session: %v\n", err)
			os.Exit(1)
		}
	}
}

func runFzf(items []string, prompt string) (string, error) {
	cmd := exec.Command("fzf", "--height=40%", "--layout=reverse", "--border", "--prompt="+prompt)
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	go func() {
		defer stdin.Close()
		for _, item := range items {
			io.WriteString(stdin, item+"\n")
		}
	}()

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func selectAction(path string, actions []api.Action) (api.Action, error) {
	var names []string
	for _, a := range actions {
		names = append(names, a.Name)
	}

	header := fmt.Sprintf("Select Action for %s", filepath.Base(path))

	cmd := exec.Command("fzf", "--height=20%", "--layout=reverse", "--border", "--header="+header, "--prompt=Action ➜ ")
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return api.Action{}, err
	}

	go func() {
		defer stdin.Close()
		for _, n := range names {
			io.WriteString(stdin, n+"\n")
		}
	}()

	output, err := cmd.Output()
	if err != nil {
		return api.Action{}, err
	}
	selectedName := strings.TrimSpace(string(output))

	for _, a := range actions {
		if a.Name == selectedName {
			return a, nil
		}
	}
	return api.Action{}, fmt.Errorf("action not found")
}
