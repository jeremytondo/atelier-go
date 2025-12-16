package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

// GetPIDFilePath returns the path to the PID file.
// It prefers XDG_RUNTIME_DIR, but falls back to a local state directory.
func GetPIDFilePath() string {
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir != "" {
		return filepath.Join(runtimeDir, "atelier-go.pid")
	}

	// Fallback
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "state", "atelier-go", "atelier-go.pid")
}

// WritePIDFile writes the current process ID to the PID file.
func WritePIDFile(pid int) error {
	path := GetPIDFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strconv.Itoa(pid)), 0600)
}

// ReadPIDFile reads the PID from the PID file.
func ReadPIDFile() (int, error) {
	path := GetPIDFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(data))
}

// RemovePIDFile removes the PID file.
func RemovePIDFile() error {
	return os.Remove(GetPIDFilePath())
}

// IsProcessRunning checks if a process with the given PID is running.
func IsProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Sending signal 0 is a way to check if the process exists without killing it.
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// StopDaemon stops the background process identified by the PID file.
// It returns true if a process was stopped, false if no process was running.
func StopDaemon() (bool, error) {
	pid, err := ReadPIDFile()
	if err != nil {
		return false, nil // No PID file, nothing to stop
	}

	if !IsProcessRunning(pid) {
		RemovePIDFile()
		return false, nil // Process not running
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		RemovePIDFile()
		return false, nil // Process not found
	}

	if err := proc.Kill(); err != nil {
		// If we can't kill it, maybe it's already dead or we don't have permission.
		// Check if it's still running.
		if IsProcessRunning(pid) {
			return false, fmt.Errorf("failed to kill process %d: %w", pid, err)
		}
	}

	// Wait briefly for process to terminate
	time.Sleep(500 * time.Millisecond)
	RemovePIDFile()
	return true, nil
}

// StartDetached starts the current executable with the given arguments as a detached process.
// It writes the PID file and returns the new PID.
func StartDetached(args ...string) (int, error) {
	// Check if already running
	pid, err := ReadPIDFile()
	if err == nil {
		if IsProcessRunning(pid) {
			return pid, fmt.Errorf("process already running (PID: %d)", pid)
		}
		// Stale PID file
		RemovePIDFile()
	}

	// Find the executable
	exe, err := os.Executable()
	if err != nil {
		return 0, fmt.Errorf("failed to determine executable path: %w", err)
	}

	// Start the process detached
	cmd := exec.Command(exe, args...)

	// Detach I/O
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start process: %w", err)
	}

	if err := WritePIDFile(cmd.Process.Pid); err != nil {
		cmd.Process.Kill()
		return 0, fmt.Errorf("failed to write PID file: %w", err)
	}

	return cmd.Process.Pid, nil
}
