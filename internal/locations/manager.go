package locations

import (
	"context"
	"fmt"
	"sync"
)

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
				// We log errors but don't fail completely if one provider fails?
				// For now, let's just store the error.
				errors[index] = err
				return
			}
			results[index] = locs
		}(i, p)
	}

	wg.Wait()

	// Check for errors, maybe return the first one?
	// Or maybe just log them and continue with what we have?
	// The prompt implies we should be robust.
	// Let's return an error if ALL providers fail, otherwise return partial results?
	// Actually, simplicity first. If a provider fails, we return the error.
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
