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
	projects         []config.Project
	defaultActions   []config.Action
	rootShellDefault bool
}

// NewProjectProvider creates a new ProjectProvider.
func NewProjectProvider(projects []config.Project, defaultActions []config.Action, rootShellDefault bool) *ProjectProvider {
	return &ProjectProvider{
		projects:         projects,
		defaultActions:   defaultActions,
		rootShellDefault: rootShellDefault,
	}
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

		// Expand path
		expandedPath, err := utils.ExpandPath(proj.Path)
		if err != nil {
			expandedPath = proj.Path
		}

		// Validate path exists
		if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
			// Skip projects that don't exist on the current machine
			continue
		}

		actions := proj.Actions
		if proj.UseDefaultActions() {
			actions = config.MergeActions(p.defaultActions, proj.Actions)
		}

		shellDefault := proj.GetShellDefault(p.rootShellDefault)
		actions = BuildActionsWithShell(actions, shellDefault)

		locations = append(locations, Location{
			Name:    proj.Name,
			Path:    filepath.Clean(expandedPath),
			Source:  p.Name(),
			Actions: actions,
		})
	}

	return locations, nil
}
