package api

import (
	"net/http"
)

// Config holds the configuration for the API server.
type Config struct {
	Actions []Action
}

// Server holds the state and dependencies for the API server.
type Server struct {
	config Config
}

// NewServer creates a new Server instance with the given configuration.
func NewServer(config Config) *Server {
	return &Server{config: config}
}

// Routes returns the HTTP handler for the server's API routes.
// It applies the given middleware to all routes.
func (s *Server) Routes(middleware func(http.Handler) http.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/health", middleware(http.HandlerFunc(s.HealthHandler)))
	mux.Handle("/api/locations", middleware(http.HandlerFunc(s.LocationsHandler)))
	mux.Handle("/api/actions", middleware(http.HandlerFunc(s.ActionsHandler)))

	return mux
}
