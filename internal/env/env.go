// Package env handles environment bootstrapping and management.
package env

import (
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

// ConfigurePath ensures that the directory containing the current executable
// and other common user binary locations are in the PATH.
func ConfigurePath() {
	path := os.Getenv("PATH")
	paths := filepath.SplitList(path)
	pathMap := make(map[string]bool)
	for _, p := range paths {
		pathMap[p] = true
	}

	var newPaths []string

	// 1. Add the directory of the current executable
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		if !pathMap[exeDir] {
			newPaths = append(newPaths, exeDir)
			pathMap[exeDir] = true
		}
	}

	// 2. Add common user binary locations
	home, _ := os.UserHomeDir()
	if home != "" {
		locals := []string{
			filepath.Join(home, ".local", "bin"),
			filepath.Join(home, "bin"),
			filepath.Join(home, "go", "bin"),
		}
		for _, p := range locals {
			if !pathMap[p] {
				if _, err := os.Stat(p); err == nil {
					newPaths = append(newPaths, p)
					pathMap[p] = true
				}
			}
		}
	}

	if len(newPaths) > 0 {
		newPath := strings.Join(append(newPaths, paths...), string(os.PathListSeparator))
		os.Setenv("PATH", newPath)
	}
}

// BuildInteractiveWrapper wraps a command to run inside an interactive login shell,
// ensuring that the user's full environment (profiles, rc files) is loaded.
func BuildInteractiveWrapper(shell, cmd string) []string {
	if cmd == "" {
		// Just launch the shell as a login shell
		return []string{shell, "-l", "-i"}
	}
	// Wrap command in interactive login shell
	return []string{shell, "-l", "-i", "-c", cmd}
}

// DetectShell identifies the user's login shell.
func DetectShell() string {
	shell := os.Getenv("SHELL")
	if shell != "" && shell != "/bin/sh" {
		return shell
	}

	if loginShell := getLoginShell(); loginShell != "" {
		if _, err := os.Stat(loginShell); err == nil {
			return loginShell
		}
	}

	for _, p := range []string{
		"/bin/zsh", "/usr/bin/zsh", "/usr/local/bin/zsh", "/opt/homebrew/bin/zsh",
		"/bin/bash", "/usr/bin/bash", "/usr/local/bin/bash",
		"/usr/bin/fish", "/usr/local/bin/fish", "/opt/homebrew/bin/fish",
	} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return "/bin/sh"
}

func getLoginShell() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}

	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("dscl", ".", "-read", "/Users/"+u.Username, "UserShell").Output()
		if err == nil {
			parts := strings.Split(string(out), ":")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	case "linux":
		out, err := exec.Command("getent", "passwd", u.Username).Output()
		if err == nil {
			parts := strings.Split(strings.TrimSpace(string(out)), ":")
			if len(parts) >= 7 {
				return parts[6]
			}
		}
	}
	return ""
}
