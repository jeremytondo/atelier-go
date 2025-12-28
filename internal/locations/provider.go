package locations

import "context"

// Provider defines the interface for location sources.
type Provider interface {
	Name() string
	Fetch(ctx context.Context) ([]Location, error)
}
