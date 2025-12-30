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
	headers := []string{"SOURCE", "NAME", "PATH", "ACTIONS"}
	var rows [][]string

	for _, loc := range locs {
		actionCount := len(loc.Actions)
		actionStr := "-"
		if actionCount > 0 {
			actionStr = fmt.Sprintf("%d", actionCount)
		}
		rows = append(rows, []string{loc.Source, loc.Name, loc.Path, actionStr})
	}

	return utils.RenderTable(w, headers, rows)
}
