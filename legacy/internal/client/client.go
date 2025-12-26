package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"atelier-go/internal/api"
	"atelier-go/internal/shell"
	"atelier-go/internal/system"
)

type Client struct {
	Host  string
	Port  int
	Token string
}

// New creates a new Client instance.
func New(host string, port int, token string) *Client {
	return &Client{
		Host:  host,
		Port:  port,
		Token: token,
	}
}

// FetchLocations retrieves the list of locations (sessions, projects, paths) from the server.
func (c *Client) FetchLocations(filter string) (*api.LocationsResponse, error) {
	url := fmt.Sprintf("http://%s:%d/api/locations?filter=%s", c.Host, c.Port, filter)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

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

// FetchActions retrieves the list of available actions for a given path.
func (c *Client) FetchActions(path string) (*api.ActionsResponse, error) {
	urlBase := fmt.Sprintf("http://%s:%d/api/actions", c.Host, c.Port)
	req, err := http.NewRequest("GET", urlBase, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("path", path)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Authorization", "Bearer "+c.Token)

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

// Attach connects to an existing session.
func (c *Client) Attach(sessionName string) error {
	bin, args := shell.BuildAttachArgs(c.Host, sessionName)

	if system.IsLocal(c.Host) {
		fmt.Printf("\033]2;%s\007", sessionName)
	}

	if bin == "ssh" {
		return runSSH(args)
	}
	cmd := exec.Command(bin, args...)
	return runInteractive(cmd)
}

// Start creates and attaches to a new session.
func (c *Client) Start(path, name string, action api.Action, isProject bool) error {
	// Prepare session info (ID)
	info := shell.PrepareSessionInfo(path, name, action.Name, isProject)

	// Build the command arguments
	bin, args := shell.BuildStartArgs(c.Host, path, info, action.Command)

	// Update local title if running locally
	if system.IsLocal(c.Host) && info.ID != "" {
		fmt.Printf("\033]2;%s\007", info.ID)
	}

	if bin == "ssh" {
		return runSSH(args)
	}
	cmd := exec.Command(bin, args...)
	return runInteractive(cmd)
}

func runSSH(args []string) error {
	fmt.Printf("Connecting: ssh %s\n", strings.Join(args, " "))
	cmd := exec.Command("ssh", args...)
	return runInteractive(cmd)
}

func runInteractive(cmd *exec.Cmd) error {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr
		}
		return err
	}
	return nil
}
