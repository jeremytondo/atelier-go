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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	iconSession = " "
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Connect to the Atelier daemon",
	Run: func(cmd *cobra.Command, args []string) {
		runClient()
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)
}

func runClient() {
	host := viper.GetString("host")
	if host == "" {
		host = "localhost"
	}
	port := viper.GetInt("port")

	// 1. Load Token
	tokenPath := os.ExpandEnv("$HOME/.config/atelier/token")
	token, err := auth.LoadOrCreateToken(tokenPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading token: %v\n", err)
		os.Exit(1)
	}

	// 2. Fetch Locations
	url := fmt.Sprintf("http://%s:%d/api/locations", host, port)
	locations, err := fetchLocations(url, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to Atelier Daemon at %s: %v\n", url, err)
		fmt.Fprintf(os.Stderr, "Is the server running? (atelier-go server)\n")
		os.Exit(1)
	}

	// 3. Prepare List for FZF
	var options []string
	// Add Sessions first with icon
	for _, s := range locations.Sessions {
		options = append(options, iconSession+" "+s)
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
	if strings.HasPrefix(selection, iconSession) {
		// Attach to existing session
		sessionName := strings.TrimSpace(strings.TrimPrefix(selection, iconSession))
		connectToSession(host, sessionName)
	} else {
		// Create new session at path
		path := selection
		action, err := selectAction(path)
		if err != nil {
			os.Exit(0)
		}
		createNewSession(host, path, action)
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

func selectAction(path string) (string, error) {
	actions := []string{"edit", "shell", "opencode"}
	header := fmt.Sprintf("Select Action for %s", filepath.Base(path))

	// Basic fzf call for actions
	cmd := exec.Command("fzf", "--height=20%", "--layout=reverse", "--border", "--header="+header, "--prompt=Action ➜ ")
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	go func() {
		defer stdin.Close()
		for _, a := range actions {
			io.WriteString(stdin, a+"\n")
		}
	}()

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func connectToSession(host, sessionName string) {
	// ssh -t <HOST> shpool attach -f <SESSION_NAME>
	// Note: If host is "localhost" or "0.0.0.0", we might want to skip SSH?
	// But the prompt says "Construct the final SSH command". So we stick to that.

	// Quote sessionName to prevent remote shell globbing
	quotedSession := fmt.Sprintf("'%s'", sessionName)
	sshArgs := []string{"-t", host, "shpool", "attach", "-f", quotedSession}
	runSSH(sshArgs)
}

func createNewSession(host, path, action string) {
	// Map action to command
	var cmd string
	switch action {
	case "edit":
		cmd = "nvim ." // simplified, assuming nvim exists on remote as per previous scripts
	case "shell":
		cmd = "${SHELL:-/bin/bash} -l"
	case "opencode":
		cmd = "opencode"
	default:
		cmd = "${SHELL:-/bin/bash}"
	}

	// Construct session name: [path:action]
	sessionName := fmt.Sprintf("[%s:%s]", path, action)

	// ssh -t <HOST> shpool attach --dir <PATH> --cmd <CMD> <SESSION_ID>
	// Quote arguments to prevent remote shell globbing or word splitting
	sshArgs := []string{
		"-t", host,
		"shpool", "attach",
		"--dir", fmt.Sprintf("'%s'", path),
		"--cmd", fmt.Sprintf("\"%s\"", cmd),
		fmt.Sprintf("'%s'", sessionName),
	}
	runSSH(sshArgs)
}

func runSSH(args []string) {
	fmt.Printf("Connecting: ssh %s\n", strings.Join(args, " "))

	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// If ssh fails, it usually prints to stderr.
		// We might want to handle exit codes, but for now just exit.
		// Note: ssh -t returning after session ends might return non-zero if the inner command did?
		// Or if connection closed.
		// Generally we can just exit.
		os.Exit(1)
	}
}
