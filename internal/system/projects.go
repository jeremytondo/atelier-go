package system

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type ProjectAction struct {
	Name    string `mapstructure:"name" json:"name"`
	Command string `mapstructure:"command" json:"command"`
}

type Project struct {
	Name     string          `mapstructure:"name" json:"name"`
	Location string          `mapstructure:"location" json:"location"`
	Actions  []ProjectAction `mapstructure:"actions" json:"actions"`
}

// LoadProjects reads all .toml files in the projects config directory.
func LoadProjects() ([]Project, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	projectsDir := filepath.Join(configDir, "projects")
	files, err := filepath.Glob(filepath.Join(projectsDir, "*.toml"))
	if err != nil {
		return []Project{}, nil
	}

	var projects []Project
	for _, file := range files {
		v := viper.New()
		v.SetConfigFile(file)
		if err := v.ReadInConfig(); err != nil {
			continue
		}

		var p Project
		if err := v.Unmarshal(&p); err != nil {
			continue
		}

		if p.Location != "" {
			// Expand ~ to home directory
			if strings.HasPrefix(p.Location, "~/") {
				home, err := os.UserHomeDir()
				if err == nil {
					p.Location = filepath.Join(home, p.Location[2:])
				}
			}

			// Expand environment variables in the location path
			p.Location = os.ExpandEnv(p.Location)
			projects = append(projects, p)
		}
	}
	return projects, nil
}

// GetProjectEntries returns a list of project paths.
func GetProjectEntries() ([]string, error) {
	projects, err := LoadProjects()
	if err != nil {
		return []string{}, err
	}

	var paths []string
	for _, p := range projects {
		paths = append(paths, p.Location)
	}
	return paths, nil
}

// GetProjectByPath returns the project config for a given path, or nil if not found.
func GetProjectByPath(path string) *Project {
	projects, err := LoadProjects()
	if err != nil {
		return nil
	}

	cleanPath := filepath.Clean(path)

	for _, p := range projects {
		if filepath.Clean(p.Location) == cleanPath {
			return &p
		}
	}
	return nil
}
