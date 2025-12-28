package locations

import (
	"atelier-go/internal/config"
	"atelier-go/internal/utils"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// ProjectProvider implements Provider for configured projects.
type ProjectProvider struct {
	configDir string
}

// NewProjectProvider creates a new ProjectProvider.
func NewProjectProvider(configDir string) *ProjectProvider {
	return &ProjectProvider{configDir: configDir}
}

// Name returns the provider name.
func (p *ProjectProvider) Name() string {
	return "Project"
}

// Fetch returns the list of configured projects as Locations.
func (p *ProjectProvider) Fetch(ctx context.Context) ([]Location, error) {
	projectsDir := filepath.Join(p.configDir, "projects")

	// Create directory if it doesn't exist
	if _, err := os.Stat(projectsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(projectsDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create projects directory: %w", err)
		}
	}

	files, err := filepath.Glob(filepath.Join(projectsDir, "*.toml"))
	if err != nil {
		return nil, fmt.Errorf("failed to list project files: %w", err)
	}

	var locations []Location
	for _, file := range files {
		v := viper.New()
		v.SetConfigFile(file)
		if err := v.ReadInConfig(); err != nil {
			// Skip malformed files, but maybe log it if we had a logger
			continue
		}

		var proj config.Project
		if err := v.Unmarshal(&proj); err != nil {
			continue
		}

		if proj.Path == "" {
			continue
		}

		// Expand path
		expandedPath, err := utils.ExpandPath(proj.Path)
		if err != nil {
			// Keep original if expansion fails
			expandedPath = proj.Path
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
