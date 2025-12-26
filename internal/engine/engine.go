package engine

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
	// 1. Load Config
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 2. Query Zoxide
	zoxidePaths, err := zoxide.Query()
	if err != nil {
		// Start with an empty list if zoxide fails, or return error?
		// The plan implies we should merge. If zoxide fails, maybe we just want config?
		// However, "failed to run zoxide query" is usually a real error.
		// Let's return the error to be safe, or we could log and continue.
		// Given this is a CLI tool, failing fast or returning error is usually better unless specified otherwise.
		return nil, fmt.Errorf("failed to query zoxide: %w", err)
	}

	var items []Item
	seenPaths := make(map[string]bool)

	// 3. Process Config Projects (Priority)
	for _, proj := range cfg.Projects {
		// Normalize path to ensure deduplication works correctly
		// (e.g. handling trailing slashes, though filepath.Clean does a fair job)
		cleanPath := filepath.Clean(proj.Path)

		items = append(items, Item{
			Name:    proj.Name,
			Path:    cleanPath,
			Source:  "Project",
			Actions: proj.Actions,
		})
		seenPaths[cleanPath] = true
	}

	// 4. Process Zoxide Paths
	for _, path := range zoxidePaths {
		cleanPath := filepath.Clean(path)

		// Deduplicate: If Config already has this path, skip it.
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
