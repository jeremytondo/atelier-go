package shell

import (
	"fmt"
	"os"
	"strings"

	"atelier-go/internal/system"
)

// SessionInfo holds the calculated session ID.
type SessionInfo struct {
	ID string
}

// PrepareSessionInfo calculates the session ID based on inputs.
func PrepareSessionInfo(path, name, actionName string, isProject bool) SessionInfo {
	if isProject {
		safeName := strings.ReplaceAll(strings.ToLower(name), " ", "-")
		safeActionName := strings.ReplaceAll(strings.ToLower(actionName), " ", "-")
		return SessionInfo{
			ID: fmt.Sprintf("[%s:%s]", safeName, safeActionName),
		}
	}

	safePath := strings.ReplaceAll(path, " ", "-")
	safeAction := strings.ReplaceAll(actionName, " ", "-")
	id := fmt.Sprintf("[%s:%s]", safePath, safeAction)
	return SessionInfo{
		ID: id,
	}
}

// BuildAttachArgs generates the command and arguments to attach to a session.
// It returns the executable (e.g., "ssh" or "shpool") and the arguments.
func BuildAttachArgs(host, sessionName string) (string, []string) {
	quotedSession := Quote(sessionName)

	if system.IsLocal(host) {
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
	if system.IsLocal(host) {
		shell = os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/bash"
		}
	} else {
		// For remote, we rely on remote shell expansion
		shell = "${SHELL:-/bin/bash}"
	}

	// 2. Build the inner shell command
	// e.g. $SHELL -c 'command'
	// We no longer force -l -i; we rely on the action command to define its environment.
	// However, we still wrap it in a shell -c to handle parsing/arguments correctly.
	if actionCmd == "" {
		actionCmd = shell
	}
	finalCmd := fmt.Sprintf("%s -c %s", shell, Quote(actionCmd))

	// 3. Build the shpool command
	if system.IsLocal(host) {
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

	if info.ID != "" {
		// printf "\033]2;%s\007" "ID"
		// We escape double quotes in the ID
		safeID := strings.ReplaceAll(info.ID, "\"", "\\\"")
		args = append(args, "printf", fmt.Sprintf("\"\\033]2;%s\\007\"", safeID), "&&")
	}

	args = append(args,
		"shpool", "attach",
		"--dir", quotedPath,
		"--cmd", quotedCmd,
		quotedID,
	)

	return "ssh", args
}
