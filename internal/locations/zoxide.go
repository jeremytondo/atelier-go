package locations

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// ZoxideProvider implements Provider for zoxide directories.
type ZoxideProvider struct{}

// NewZoxideProvider creates a new ZoxideProvider.
func NewZoxideProvider() *ZoxideProvider {
	return &ZoxideProvider{}
}

// Name returns the provider name.
func (z *ZoxideProvider) Name() string {
	return "Zoxide"
}

// Fetch queries zoxide for frequent directories.
func (z *ZoxideProvider) Fetch(ctx context.Context) ([]Location, error) {
	cmd := exec.CommandContext(ctx, "zoxide", "query", "-l")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run zoxide query: %w", err)
	}

	var locations []Location
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		path := strings.TrimSpace(scanner.Text())
		if path != "" {
			cleanPath := filepath.Clean(path)
			locations = append(locations, Location{
				Name:   filepath.Base(cleanPath),
				Path:   cleanPath,
				Source: z.Name(),
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse zoxide output: %w", err)
	}

	return locations, nil
}
