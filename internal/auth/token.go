package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var currentToken string

// LoadOrCreateToken ensures a token exists at the given path.
// If it doesn't, it generates one and saves it.
// Returns the token string.
func LoadOrCreateToken(path string) (string, error) {
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
			currentToken = token
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

	currentToken = token
	return token, nil
}

// GetCurrentToken returns the currently loaded token
func GetCurrentToken() string {
	return currentToken
}
