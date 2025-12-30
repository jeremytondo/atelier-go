package locations

import (
	"atelier-go/internal/config"
	"atelier-go/internal/utils"
	"context"
	"os"
	"path/filepath"
)

// ProjectProvider implements Provider for configured projects.
type ProjectProvider struct {
	projects []config.Project
}

// NewProjectProvider creates a new ProjectProvider.
func NewProjectProvider(projects []config.Project) *ProjectProvider {
	return &ProjectProvider{projects: projects}
}

// Name returns the provider name.
func (p *ProjectProvider) Name() string {
	return "Project"
}

// Fetch returns the list of configured projects as Locations.
func (p *ProjectProvider) Fetch(ctx context.Context) ([]Location, error) {
	var locations []Location

	for _, proj := range p.projects {
		if proj.Path == "" {
			continue
		}

		// 1. Expand path
		expandedPath, err := utils.ExpandPath(proj.Path)
		if err != nil {
			expandedPath = proj.Path
		}

		// 2. Validate path exists
		if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
			// Skip projects that don't exist on the current machine
			continue
		}

		locations = append(locations, Location{
			Name:    proj.Name,
			Path:    filepath.Clean(expandedPath),
			Source:  p.Name(),
			Actions: proj.Actions,
		})
	}

	return locations, nil
}
