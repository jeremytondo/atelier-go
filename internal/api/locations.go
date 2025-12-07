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

	filter := r.URL.Query().Get("filter")
	if filter == "" {
		filter = "frequent"
	}

	// Calculate which concurrent tasks to run
	runSessions := true
	runPaths := false
	pathSource := "zoxide" // "zoxide" or "fd"

	switch filter {
	case "sessions":
		runPaths = false
	case "all":
		runPaths = true
		pathSource = "fd"
	default: // "frequent"
		runPaths = true
		pathSource = "zoxide"
	}

	tasks := 0
	if runSessions {
		tasks++
	}
	if runPaths {
		tasks++
	}
	wg.Add(tasks)

	if runSessions {
		go func() {
			defer wg.Done()
			sessions, _ = system.GetSessions()
		}()
	}

	if runPaths {
		go func() {
			defer wg.Done()
			if pathSource == "fd" {
				paths, _ = system.GetAllEntries()
			} else {
				paths, _ = system.GetZoxideEntries()
			}
		}()
	}

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
