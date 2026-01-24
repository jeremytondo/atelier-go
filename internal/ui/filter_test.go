package ui

import (
	"atelier-go/internal/locations"
	"testing"
)

func TestApplyLocationFilter(t *testing.T) {
	locs := []locations.Location{
		{Name: "work", Path: "/home/user/work", Source: "Zoxide"},
		{Name: "atelier", Path: "/home/user/atelier", Source: "Project"},
		{Name: "dotfiles", Path: "/home/user/dotfiles", Source: "Zoxide"},
		{Name: "go-project", Path: "/home/user/go-project", Source: "Project"},
	}

	m := NewModel(locs)

	tests := []struct {
		name     string
		query    string
		expected []string // Names in expected order
	}{
		{
			name:  "empty query - projects first",
			query: "",
			expected: []string{
				"atelier",
				"go-project",
				"work",
				"dotfiles",
			},
		},
		{
			name:  "fuzzy query - projects first",
			query: "o",
			expected: []string{
				"go-project", // Project, matches 'o'
				"work",       // Zoxide, matches 'o'
				"dotfiles",   // Zoxide, matches 'o'
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.filterInput.SetValue(tt.query)
			m.applyLocationFilter()

			items := m.locations.Items()
			if len(items) != len(tt.expected) {
				t.Fatalf("expected %d items, got %d", len(tt.expected), len(items))
			}

			for i, name := range tt.expected {
				item := items[i].(LocationItem)
				if item.Location.Name != name {
					t.Errorf("at index %d: expected %s, got %s", i, name, item.Location.Name)
				}
			}
		})
	}
}
