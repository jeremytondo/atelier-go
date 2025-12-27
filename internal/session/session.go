// Package session defines the interface for session management.
package session

// Manager defines the behavior for attaching to workspace sessions.
type Manager interface {
	// Attach connects to an existing session or creates a new one.
	// name is the unique identifier for the session.
	// dir is the working directory for the session.
	// args are optional commands to execute within the session.
	Attach(name string, dir string, args ...string) error
}
