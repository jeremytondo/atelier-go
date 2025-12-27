package locations

import (
	"atelier-go/internal/config"
	"context"
	"path/filepath"
)

// ProjectProvider implements Provider for configured projects.
type ProjectProvider struct {
	config *config.Config
}

// NewProjectProvider creates a new ProjectProvider.
func NewProjectProvider(cfg *config.Config) *ProjectProvider {
	return &ProjectProvider{config: cfg}
}

// Name returns the provider name.
func (p *ProjectProvider) Name() string {
	return "Project"
}

// Fetch returns the list of configured projects as Locations.
func (p *ProjectProvider) Fetch(ctx context.Context) ([]Location, error) {
	var locations []Location
	for _, proj := range p.config.Projects {
		locations = append(locations, Location{
			Name:    proj.Name,
			Path:    filepath.Clean(proj.Path),
			Source:  p.Name(),
			Actions: proj.Actions,
		})
	}
	return locations, nil
}
