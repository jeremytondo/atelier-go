package ui

import (
	"atelier-go/internal/locations"
	"testing"
)

func TestNewModel_InitialSorting(t *testing.T) {
	// Mixed input: Zoxide, Project, Zoxide
	locs := []locations.Location{
		{Name: "work", Path: "/home/user/work", Source: "Zoxide"},
		{Name: "atelier", Path: "/home/user/atelier", Source: "Project"},
		{Name: "dotfiles", Path: "/home/user/dotfiles", Source: "Zoxide"},
	}

	m := NewModel(locs)

	items := m.locations.Items()
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	// Expect: Project first
	expected := []string{"atelier", "work", "dotfiles"}
	
	// Check items against expected (assuming stable sort of original zoxide items)
	for i, name := range expected {
		item := items[i].(LocationItem)
		if item.Location.Name != name {
			t.Errorf("at index %d: expected %s, got %s", i, name, item.Location.Name)
		}
	}
}

