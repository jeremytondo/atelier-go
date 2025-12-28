// Package locations handles the core logic for aggregating projects and session targets.
package locations

import (
	"atelier-go/internal/config"
	"context"
	"fmt"
	"sync"
)

// FetchOptions defines criteria for fetching locations.
type FetchOptions struct {
	IncludeProjects bool
	IncludeZoxide   bool
}

// Location represents a unified project or directory entry.
type Location struct {
	Name    string
	Path    string
	Source  string // "Project" or "Zoxide"
	Actions []config.Action
}

// Manager orchestrates location providers.
type Manager struct {
	providers []Provider
}

// NewManager creates a new Manager with the given providers.
func NewManager(providers ...Provider) *Manager {
	return &Manager{providers: providers}
}

// List discovers and returns a merged list of locations based on the provided options.
// It handles configuration loading and provider initialization internally.
func List(ctx context.Context, opts FetchOptions) ([]Location, error) {
	// 1. Get Config Directory (needed for ProjectProvider)
	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to determine config directory: %w", err)
	}

	// 2. Initialize Providers based on options
	var providers []Provider
	if opts.IncludeProjects {
		providers = append(providers, NewProjectProvider(configDir))
	}
	if opts.IncludeZoxide {
		providers = append(providers, NewZoxideProvider())
	}

	// 3. Create Manager and Fetch
	manager := NewManager(providers...)
	return manager.GetAll(ctx)
}

// GetAll returns a merged list of locations from all providers.
// It deduplicates paths, giving priority to earlier providers in the list.
func (m *Manager) GetAll(ctx context.Context) ([]Location, error) {
	var allLocations []Location
	seenPaths := make(map[string]bool)
	var mu sync.Mutex
	var wg sync.WaitGroup

	results := make([][]Location, len(m.providers))
	errors := make([]error, len(m.providers))

	for i, p := range m.providers {
		wg.Add(1)
		go func(index int, provider Provider) {
			defer wg.Done()
			locs, err := provider.Fetch(ctx)
			if err != nil {
				errors[index] = err
				return
			}
			results[index] = locs
		}(i, p)
	}

	wg.Wait()

	// Check for errors from any provider
	for _, err := range errors {
		if err != nil {
			return nil, fmt.Errorf("provider failed: %w", err)
		}
	}

	// Merge results in order
	for _, locs := range results {
		for _, loc := range locs {
			mu.Lock()
			if !seenPaths[loc.Path] {
				allLocations = append(allLocations, loc)
				seenPaths[loc.Path] = true
			}
			mu.Unlock()
		}
	}

	return allLocations, nil
}
