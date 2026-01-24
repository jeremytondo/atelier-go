package config

import (
	"fmt"
)

// Validate checks the configuration for errors.
func (c *Config) Validate() error {
	for i, p := range c.Projects {
		if p.Name == "" {
			return fmt.Errorf("project at index %d missing name", i)
		}
		if p.Path == "" {
			return fmt.Errorf("project '%s' missing path", p.Name)
		}
	}
	return nil
}
