package shell

import (
	"fmt"
	"os"
	"strings"
)

// IsLocal checks if the provided host resolves to the local machine.
func IsLocal(host string) bool {
	if host == "localhost" || host == "127.0.0.1" || host == "0.0.0.0" || host == "" {
		return true
	}
	hostname, err := os.Hostname()
	if err == nil && strings.EqualFold(host, hostname) {
		return true
	}
	return false
}

// SessionInfo holds the calculated session ID and window title.
type SessionInfo struct {
	ID    string
	Title string
}

// PrepareSessionInfo calculates the session ID and window title based on inputs.
func PrepareSessionInfo(path, name, actionName string, isProject bool) SessionInfo {
	if isProject {
		safeName := strings.ReplaceAll(strings.ToLower(name), " ", "-")
		safeActionName := strings.ReplaceAll(strings.ToLower(actionName), " ", "-")
		return SessionInfo{
			ID:    fmt.Sprintf("[%s:%s]", safeName, safeActionName),
			Title: fmt.Sprintf("%s - %s", name, actionName),
		}
	}

	safePath := strings.ReplaceAll(path, " ", "-")
	safeAction := strings.ReplaceAll(actionName, " ", "-")
	id := fmt.Sprintf("[%s:%s]", safePath, safeAction)
	return SessionInfo{
		ID:    id,
		Title: id,
	}
}

// BuildAttachArgs generates the command and arguments to attach to a session.
// It returns the executable (e.g., "ssh" or "shpool") and the arguments.
func BuildAttachArgs(host, sessionName string) (string, []string) {
	quotedSession := Quote(sessionName)

	if IsLocal(host) {
		return "shpool", []string{"attach", "-f", sessionName}
	}

	// Remote
	// Command: printf "\033]2;%s\007" <session> && shpool attach -f <session>
	// We construct it as arguments to ssh.
	// We use the quoted session name in the printf argument to be safe.
	return "ssh", []string{
		"-t", host,
		fmt.Sprintf("printf \"\\033]2;%%s\\007\" %s", quotedSession),
		"&&",
		"shpool", "attach", "-f", quotedSession,
	}
}

// BuildStartArgs generates the command and arguments to start (and attach to) a new session.
func BuildStartArgs(host, path string, info SessionInfo, actionCmd string) (string, []string) {
	// 1. Determine Shell
	shell := "$SHELL"
	if IsLocal(host) {
		shell = os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/bash"
		}
	} else {
		// For remote, we rely on remote shell expansion
		shell = "${SHELL:-/bin/bash}"
	}

	// 2. Build the inner shell command
	// e.g. $SHELL -l -i -c 'command'
	var finalCmd string
	if actionCmd == "" {
		finalCmd = fmt.Sprintf("%s -l -i", shell)
	} else {
		// Quote the action command for the -c flag
		finalCmd = fmt.Sprintf("%s -l -i -c %s", shell, Quote(actionCmd))
	}

	// 3. Build the shpool command
	if IsLocal(host) {
		return "shpool", []string{
			"attach",
			"--dir", path,
			"--cmd", finalCmd,
			info.ID,
		}
	}

	// 4. Remote SSH
	// shpool attach --dir 'path' --cmd "finalCmd" 'sessionID'
	quotedPath := Quote(path)
	quotedID := Quote(info.ID)

	// Escape double quotes in finalCmd for wrapping in double quotes for the SSH command string
	safeCmd := strings.ReplaceAll(finalCmd, "\"", "\\\"")
	quotedCmd := fmt.Sprintf("\"%s\"", safeCmd)

	args := []string{
		"-t", host,
	}

	if info.Title != "" {
		// printf "\033]2;%s\007" "Title"
		// We escape double quotes in the title
		safeTitle := strings.ReplaceAll(info.Title, "\"", "\\\"")
		args = append(args, "printf", fmt.Sprintf("\"\\033]2;%s\\007\"", safeTitle), "&&")
	}

	args = append(args,
		"shpool", "attach",
		"--dir", quotedPath,
		"--cmd", quotedCmd,
		quotedID,
	)

	return "ssh", args
}
