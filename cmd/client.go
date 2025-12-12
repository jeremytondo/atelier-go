package cmd

import (
	"fmt"
	"io"
	"maps"
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
		listMode, _ := cmd.Flags().GetBool("list")

		filter := "frequent"
		if sessions {
			filter = "sessions"
		} else if all {
			filter = "all"
		}
		runClient(filter, listMode)
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
	clientCmd.Flags().Bool("list", false, "Output raw list for fzf (internal use)")
	rootCmd.AddCommand(clientCmd)
	clientCmd.AddCommand(clientLoginCmd)
}

func runClient(filter string, listMode bool) {
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
	options, err := fetchAndFormatLocations(c, filter)
	if err != nil {
		url := fmt.Sprintf("http://%s:%d", host, port)
		fmt.Fprintf(os.Stderr, "Error connecting to Atelier Daemon at %s: %v\n", url, err)
		fmt.Fprintf(os.Stderr, "Is the server running? (atelier-go server)\n")
		os.Exit(1)
	}

	if listMode {
		for _, opt := range options {
			fmt.Println(opt)
		}
		return
	}

	if len(options) == 0 {
		// Proceed even if empty so users can switch filters using keys
	}

	// 3. Select Location via FZF
	selection, err := runFzfWithBindings(options, filter)
	if err != nil {
		// User likely cancelled
		os.Exit(0)
	}

	// 4. Handle Selection
	// Split by tab to get display name and real path
	parts := strings.Split(selection, "\t")
	if len(parts) < 2 {
		parts = []string{selection, selection}
	}
	dispName := parts[0]
	realPath := parts[1]

	if strings.HasPrefix(dispName, iconSession) {
		// Attach to existing session
		if err := c.Attach(realPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error attaching to session: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Create new session at path

		// Fetch available actions from server
		actionsResp, err := c.FetchActions(realPath)
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

		action, err := selectAction(realPath, actions)
		if err != nil {
			os.Exit(0)
		}

		// Determine the name to use (project name or fallback to path base)
		name := filepath.Base(realPath)

		if nameStr, ok := strings.CutPrefix(dispName, iconProject); ok {
			name = strings.TrimSpace(nameStr)
		}

		if err := c.Start(realPath, name, action, actionsResp != nil && actionsResp.IsProject); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting session: %v\n", err)
			os.Exit(1)
		}
	}
}

func fetchAndFormatLocations(c *client.Client, filter string) ([]string, error) {
	locations, err := c.FetchLocations(filter)
	if err != nil {
		return nil, err
	}

	var options []string

	// Add Sessions first with icon
	for _, s := range locations.Sessions {
		// Format: Icon SessionName \t SessionName
		options = append(options, fmt.Sprintf("%s %s\t%s", iconSession, s, s))
	}
	// Add Projects
	for _, p := range locations.Projects {
		// Use project name for display
		dispName := p.Name
		if dispName == "" {
			dispName = filepath.Base(p.Location)
		}
		// Format: Icon ProjectName \t ProjectLocation
		options = append(options, fmt.Sprintf("%s %s\t%s", iconProject, dispName, p.Location))
	}
	// Add Paths
	for _, path := range locations.Paths {
		// Format: Path \t Path
		options = append(options, fmt.Sprintf("%s\t%s", path, path))
	}

	return options, nil
}

func runFzfWithBindings(items []string, currentFilter string) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		exe = "atelier-go"
	}

	// Default bindings
	bindings := map[string]string{
		"sessions": "ctrl-s",
		"all":      "ctrl-a",
		"frequent": "ctrl-f",
	}

	// Override with config
	if viper.IsSet("keys") {
		keys := viper.GetStringMapString("keys")
		maps.Copy(bindings, keys)
	}

	var bindArgs []string
	var headerParts []string

	// Define order for header
	order := []string{"sessions", "frequent", "all"}

	for _, mode := range order {
		key, ok := bindings[mode]
		if !ok {
			continue
		}

		// Determine flag
		flag := ""
		if mode == "sessions" {
			flag = " --sessions"
		} else if mode == "all" {
			flag = " --all"
		}

		label := toTitle(mode)

		// Construct reload command
		// e.g. /path/to/atelier-go client --list --sessions
		cmdStr := fmt.Sprintf("%s client --list%s", exe, flag)

		// Construct bind string
		// e.g. ctrl-s:reload(...)+change-prompt(Sessions ➜ )
		bind := fmt.Sprintf("%s:reload(%s)+change-prompt(%s ➜ )", key, cmdStr, label)
		bindArgs = append(bindArgs, "--bind", bind)

		headerParts = append(headerParts, fmt.Sprintf("%s: %s", strings.ToUpper(key), label))
	}

	header := strings.Join(headerParts, " | ")
	prompt := toTitle(currentFilter) + " ➜ "

	cmd := exec.Command("fzf",
		"--height=40%",
		"--layout=reverse",
		"--border",
		"--prompt="+prompt,
		"--delimiter=\t",
		"--with-nth=1",
		"--header="+header,
		"--header-first",
	)

	// Add bind args
	cmd.Args = append(cmd.Args, bindArgs...)

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

func toTitle(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
