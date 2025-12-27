// Package locations handles the core logic for aggregating projects and session targets.
package locations

import (
	"fmt"
	"path/filepath"

	"atelier-go/internal/config"
	"atelier-go/internal/zoxide"
)

// Item represents a unified project or directory entry.
type Item struct {
	Name    string
	Path    string
	Source  string // "Project" or "Zoxide"
	Actions []config.Action
}

// Fetch retrieves items from configuration and zoxide, merging and deduplicating them.
// Configured projects take precedence over zoxide entries.
func Fetch() ([]Item, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	zoxidePaths, err := zoxide.Query()
	if err != nil {
		return nil, fmt.Errorf("failed to query zoxide: %w", err)
	}

	var items []Item
	seenPaths := make(map[string]bool)

	// Process configured projects first (priority)
	for _, proj := range cfg.Projects {
		cleanPath := filepath.Clean(proj.Path)

		items = append(items, Item{
			Name:    proj.Name,
			Path:    cleanPath,
			Source:  "Project",
			Actions: proj.Actions,
		})
		seenPaths[cleanPath] = true
	}

	// Append zoxide paths, avoiding duplicates
	for _, path := range zoxidePaths {
		cleanPath := filepath.Clean(path)

		if seenPaths[cleanPath] {
			continue
		}

		items = append(items, Item{
			Name:   filepath.Base(cleanPath),
			Path:   cleanPath,
			Source: "Zoxide",
		})
		seenPaths[cleanPath] = true
	}

	return items, nil
}
