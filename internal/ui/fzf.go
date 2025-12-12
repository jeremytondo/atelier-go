package ui

import (
	"atelier-go/internal/api"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func RunFzfWithBindings(items []string, currentFilter string) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		exe = "atelier-go"
	}

	// Default bindings
	bindings := map[string]string{
		api.FilterSessions: "ctrl-s",
		api.FilterProjects: "ctrl-p",
		api.FilterAll:      "ctrl-a",
		api.FilterFrequent: "ctrl-f",
	}

	// Override with config
	if viper.IsSet("keys") {
		keys := viper.GetStringMapString("keys")
		maps.Copy(bindings, keys)
	}

	var bindArgs []string
	var headerParts []string

	// Define order for header
	order := []string{api.FilterSessions, api.FilterProjects, api.FilterFrequent, api.FilterAll}

	for _, mode := range order {
		key, ok := bindings[mode]
		if !ok {
			continue
		}

		// Determine flag
		flag := ""
		switch mode {
		case api.FilterSessions:
			flag = " --sessions"
		case api.FilterProjects:
			flag = " --projects"
		case api.FilterAll:
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

func SelectAction(path string, actions []api.Action) (api.Action, error) {
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
