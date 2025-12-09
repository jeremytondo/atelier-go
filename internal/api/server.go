package api

import (
	"net/http"
)

type Config struct {
	Actions []Action
}

type Server struct {
	config Config
}

func NewServer(config Config) *Server {
	return &Server{config: config}
}

func (s *Server) Routes(middleware func(http.Handler) http.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/health", middleware(http.HandlerFunc(s.HealthHandler)))
	mux.Handle("/api/locations", middleware(http.HandlerFunc(s.LocationsHandler)))
	mux.Handle("/api/actions", middleware(http.HandlerFunc(s.ActionsHandler)))

	return mux
}
