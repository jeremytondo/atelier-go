package api

import (
	"encoding/json"
	"net/http"
	"sync"

	"atelier-go/internal/system"
)

type LocationsResponse struct {
	Sessions []string `json:"sessions"`
	Paths    []string `json:"paths"`
}

func LocationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var wg sync.WaitGroup
	var sessions []string
	var paths []string

	// We don't really need to handle errors here since the system functions swallow them gracefully
	// returning empty lists.

	wg.Add(2)

	go func() {
		defer wg.Done()
		sessions, _ = system.GetSessions()
	}()

	go func() {
		defer wg.Done()
		paths, _ = system.GetZoxideEntries()
	}()

	wg.Wait()

	// Ensure non-nil slices for JSON
	if sessions == nil {
		sessions = []string{}
	}
	if paths == nil {
		paths = []string{}
	}

	response := LocationsResponse{
		Sessions: sessions,
		Paths:    paths,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
