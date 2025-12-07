package system

import (
	"os"
	"path/filepath"
	"strconv"
	"syscall"
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
