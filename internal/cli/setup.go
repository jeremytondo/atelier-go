package cli

import (
	"atelier-go/internal/config"
	"atelier-go/internal/locations"
	"fmt"
)

// setupLocationManager creates a locations.Manager based on the provided flags.
func setupLocationManager(includeProjects, includeZoxide bool) (*locations.Manager, error) {
	// Default to showing both if neither is specified
	if !includeProjects && !includeZoxide {
		includeProjects = true
		includeZoxide = true
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	var providers []locations.Provider
	if includeProjects {
		providers = append(providers, locations.NewProjectProvider(cfg.Projects))
	}
	if includeZoxide {
		providers = append(providers, locations.NewZoxideProvider())
	}

	return locations.NewManager(providers...), nil
}
