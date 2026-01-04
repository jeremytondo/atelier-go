package locations

import (
	"atelier-go/internal/config"
	"atelier-go/internal/utils"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// ZoxideProvider implements Provider for zoxide directories.
type ZoxideProvider struct {
	defaultActions []config.Action
	shellDefault   bool
}

// NewZoxideProvider creates a new ZoxideProvider.
func NewZoxideProvider(defaultActions []config.Action, shellDefault bool) *ZoxideProvider {
	return &ZoxideProvider{
		defaultActions: defaultActions,
		shellDefault:   shellDefault,
	}
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
			// Canonicalize path (fixes case sensitivity on macOS)
			if canonical, err := utils.GetCanonicalPath(cleanPath); err == nil {
				cleanPath = canonical
			}

			locations = append(locations, Location{
				Name:    filepath.Base(cleanPath),
				Path:    cleanPath,
				Source:  z.Name(),
				Actions: BuildActionsWithShell(z.defaultActions, z.shellDefault),
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse zoxide output: %w", err)
	}

	return locations, nil
}
