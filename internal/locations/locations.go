// Package locations handles the core logic for aggregating projects and session targets.
package locations

import (
	"context"
	"fmt"
	"io"
	"sync"

	"atelier-go/internal/config"
	"atelier-go/internal/utils"
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
	// Default to showing both if neither is specified
	if !opts.IncludeProjects && !opts.IncludeZoxide {
		opts.IncludeProjects = true
		opts.IncludeZoxide = true
	}

	// Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize Providers based on options
	var providers []Provider
	if opts.IncludeProjects {
		providers = append(providers, NewProjectProvider(cfg.Projects))
	}
	if opts.IncludeZoxide {
		providers = append(providers, NewZoxideProvider())
	}

	// Create Manager and Fetch
	manager := NewManager(providers...)
	return manager.GetAll(ctx)
}

// GetAll returns a merged list of locations from all providers.
// It deduplicates paths, giving priority to earlier providers in the list.
func (m *Manager) GetAll(ctx context.Context) ([]Location, error) {
	var allLocations []Location
	seenPaths := make(map[string]bool)
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
			if !seenPaths[loc.Path] {
				allLocations = append(allLocations, loc)
				seenPaths[loc.Path] = true
			}
		}
	}

	return allLocations, nil
}

// PrintTable formats and prints the locations to the provided writer in a table format.
func PrintTable(w io.Writer, locs []Location) error {
	tw := utils.NewTableWriter(w)
	if _, err := fmt.Fprintln(tw, "SOURCE\tNAME\tPATH\tACTIONS"); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	for _, loc := range locs {
		actionCount := len(loc.Actions)
		actionStr := "-"
		if actionCount > 0 {
			actionStr = fmt.Sprintf("%d", actionCount)
		}

		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", loc.Source, loc.Name, loc.Path, actionStr); err != nil {
			return fmt.Errorf("error writing row: %w", err)
		}
	}
	if err := tw.Flush(); err != nil {
		return fmt.Errorf("error flushing writer: %w", err)
	}
	return nil
}
