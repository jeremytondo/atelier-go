package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"atelier-go/internal/api"
	"atelier-go/internal/auth"
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
	Short: "Connect to the Atelier daemon",
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

	// 2. Fetch Locations
	url := fmt.Sprintf("http://%s:%d/api/locations?filter=%s", host, port, filter)
	locations, err := fetchLocations(url, token)
	if err != nil {
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
		connectToSession(host, strings.TrimSpace(sessionName))
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
		urlActions := fmt.Sprintf("http://%s:%d/api/actions", host, port)
		actionsResp, err := fetchActions(urlActions, token, path)
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

		createNewSession(host, path, name, action, actionsResp != nil && actionsResp.IsProject)
	}
}

func fetchLocations(url, token string) (*api.LocationsResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var locs api.LocationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&locs); err != nil {
		return nil, err
	}
	return &locs, nil
}

func fetchActions(urlBase, token, path string) (*api.ActionsResponse, error) {
	req, err := http.NewRequest("GET", urlBase, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("path", path)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var actions api.ActionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&actions); err != nil {
		return nil, err
	}
	return &actions, nil
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

func connectToSession(host, sessionName string) {
	if isLocal(host) {
		fmt.Printf("\033]2;%s\007", sessionName)
		cmd := exec.Command("shpool", "attach", "-f", sessionName)
		runInteractive(cmd)
		return
	}

	// Quote sessionName to prevent remote shell globbing
	quotedSession := fmt.Sprintf("'%s'", sessionName)

	// We prepend the printf command to update the window title before attaching
	// The format is: printf "\033]2;%s\007" "sessionName"
	sshArgs := []string{
		"-t", host,
		"printf", "\"\\033]2;%s\\007\"", quotedSession,
		"&&",
		"shpool", "attach", "-f", quotedSession,
	}
	runSSH(sshArgs)
}

func createNewSession(host, path, name string, action api.Action, isProject bool) {
	// Sanitize session name: lowercase, spaces -> dashes
	// Window Title: Project Name - Action Name (original case)

	var sessionID string
	var windowTitle string

	if isProject {
		// Session ID: [project-name:action-name] (sanitized)
		safeName := strings.ReplaceAll(strings.ToLower(name), " ", "-")
		safeActionName := strings.ReplaceAll(strings.ToLower(action.Name), " ", "-")
		sessionID = fmt.Sprintf("[%s:%s]", safeName, safeActionName)

		// Window Title: Project Name - Action Name
		windowTitle = fmt.Sprintf("%s - %s", name, action.Name)
	} else {
		// Standard behavior
		// Session ID: [path:action]
		safePath := strings.ReplaceAll(path, " ", "-")
		safeAction := strings.ReplaceAll(action.Name, " ", "-")
		sessionID = fmt.Sprintf("[%s:%s]", safePath, safeAction)

		windowTitle = sessionID
	}

	cmdStr := action.Command
	if cmdStr == "" {
		cmdStr = "${SHELL:-/bin/bash}"
	} else if isLocal(host) {
		cmdStr = os.ExpandEnv(cmdStr)
	} else {
		// Remote: ensure shell default if empty
		if cmdStr == "" {
			cmdStr = "${SHELL:-/bin/bash}"
		}
	}

	if isLocal(host) {
		if windowTitle != "" {
			fmt.Printf("\033]2;%s\007", windowTitle)
		}

		cmd := exec.Command("shpool", "attach",
			"--dir", path,
			"--cmd", cmdStr,
			sessionID,
		)
		runInteractive(cmd)
		return
	}

	quotedSessionID := fmt.Sprintf("'%s'", sessionID)

	// Remote
	// ssh -t <HOST> printf "\033]2;%s\007" <TITLE> && shpool attach ... <SESSION_ID>
	sshArgs := []string{
		"-t", host,
	}

	if windowTitle != "" {
		// printf "\033]2;%s\007" "Title"
		// We need to be careful with quoting for the remote shell.
		sshArgs = append(sshArgs, "printf", fmt.Sprintf("\"\\033]2;%s\\007\"", windowTitle), "&&")
	}

	sshArgs = append(sshArgs,
		"shpool", "attach",
		"--dir", fmt.Sprintf("'%s'", path),
		"--cmd", fmt.Sprintf("\"%s\"", cmdStr),
		quotedSessionID,
	)

	runSSH(sshArgs)
}

func runSSH(args []string) {
	fmt.Printf("Connecting: ssh %s\n", strings.Join(args, " "))

	cmd := exec.Command("ssh", args...)
	runInteractive(cmd)
}

func isLocal(host string) bool {
	if host == "localhost" || host == "127.0.0.1" || host == "0.0.0.0" {
		return true
	}
	hostname, err := os.Hostname()
	if err == nil && strings.EqualFold(host, hostname) {
		return true
	}
	return false
}

func runInteractive(cmd *exec.Cmd) {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
}
