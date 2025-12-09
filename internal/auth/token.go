package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Authenticator handles token-based authentication.
type Authenticator struct {
	token string
}

// NewAuthenticator creates a new Authenticator with the given token.
func NewAuthenticator(token string) *Authenticator {
	return &Authenticator{token: token}
}

// GetDefaultTokenPath returns the XDG-compliant path for the token file
func GetDefaultTokenPath() (string, error) {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, "atelier-go", "token"), nil
}

// LoadOrCreateToken ensures a token exists at the given path.
// If ATELIER_TOKEN env var is set, it uses that.
// If it doesn't, it generates one and saves it to path.
// Returns the token string.
func LoadOrCreateToken(path string) (string, error) {
	// Check environment variable first
	if envToken := os.Getenv("ATELIER_TOKEN"); envToken != "" {
		return envToken, nil
	}

	// Expand ~ if present (simple handling for now, assuming HOME env var is set)
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home dir: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	// Try to read existing token
	content, err := os.ReadFile(path)
	if err == nil {
		token := strings.TrimSpace(string(content))
		if token != "" {
			return token, nil
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to read token file: %w", err)
	}

	// Generate new token
	bytes := make([]byte, 16) // 16 bytes = 32 hex chars
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	token := hex.EncodeToString(bytes)

	// Write to file
	if err := os.WriteFile(path, []byte(token), 0600); err != nil {
		return "", fmt.Errorf("failed to write token file: %w", err)
	}

	return token, nil
}

// SaveToken writes the given token to the specified path, creating directories if needed.
func SaveToken(path, token string) error {
	// Expand ~ if present
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home dir: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, []byte(token), 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}
