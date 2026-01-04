// Package locations handles the core logic for aggregating projects and session targets.
package locations

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"atelier-go/internal/config"
	"atelier-go/internal/utils"

	"github.com/sahilm/fuzzy"
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

// locationSource implements fuzzy.Source for location matching.
type locationSource []Location

func (s locationSource) String(i int) string { return s[i].Name }
func (s locationSource) Len() int            { return len(s) }

// Find searches for a location by name. It first tries an exact case-insensitive
// match, then falls back to fuzzy matching if no exact match is found.
func (m *Manager) Find(ctx context.Context, name string) (*Location, error) {
	locs, err := m.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Try exact match first
	for _, loc := range locs {
		if strings.EqualFold(loc.Name, name) {
			return &loc, nil
		}
	}

	// Fallback to fuzzy match
	matches := fuzzy.FindFrom(name, locationSource(locs))
	if len(matches) > 0 {
		return &locs[matches[0].Index], nil
	}

	return nil, fmt.Errorf("location %q not found", name)
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

// BuildActionsWithShell constructs the final action list, positioning "Shell"
// correctly based on the shellDefault setting. It ensures no duplicate "Shell" action
// and avoids mutating the input slice.
func BuildActionsWithShell(actions []config.Action, shellDefault bool) []config.Action {
	if len(actions) == 0 {
		return nil
	}

	// Check if "Shell" already exists (case-insensitive)
	hasShell := false
	for _, a := range actions {
		if strings.EqualFold(a.Name, "shell") {
			hasShell = true
			break
		}
	}

	if hasShell {
		// If it exists, we just return a copy of the input slice to avoid mutation
		merged := make([]config.Action, len(actions))
		copy(merged, actions)
		// If shellDefault is true, we should move it to the front if it's not there
		if shellDefault && !strings.EqualFold(merged[0].Name, "shell") {
			var shellIdx int
			for i, a := range merged {
				if strings.EqualFold(a.Name, "shell") {
					shellIdx = i
					break
				}
			}
			shell := merged[shellIdx]
			// Remove from current position and prepend
			merged = append(merged[:shellIdx], merged[shellIdx+1:]...)
			merged = append([]config.Action{shell}, merged...)
		}
		return merged
	}

	shellAction := config.Action{Name: "Shell", Command: ""}

	if shellDefault {
		merged := make([]config.Action, 0, len(actions)+1)
		merged = append(merged, shellAction)
		merged = append(merged, actions...)
		return merged
	}

	merged := make([]config.Action, 0, len(actions)+1)
	merged = append(merged, actions...)
	merged = append(merged, shellAction)
	return merged
}
