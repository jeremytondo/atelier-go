package api

import (
	"encoding/json"
	"net/http"

	"atelier-go/internal/system"
)

type Action struct {
	Name    string `json:"name" mapstructure:"name"`
	Command string `json:"command" mapstructure:"command"`
}

type ActionsResponse struct {
	Actions   []Action `json:"actions"`
	IsProject bool     `json:"is_project"`
}

func (s *Server) ActionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Copy configured actions
	actions := make([]Action, len(s.config.Actions))
	copy(actions, s.config.Actions)

	// Ensure default shell action exists
	hasShell := false
	for _, a := range actions {
		if a.Name == "shell" {
			hasShell = true
			break
		}
	}

	if !hasShell {
		actions = append(actions, Action{
			Name:    "shell",
			Command: "$SHELL -l",
		})
	}

	// Check for project context
	path := r.URL.Query().Get("path")
	isProject := false
	if path != "" {
		if proj := system.GetProjectByPath(path); proj != nil {
			isProject = true
			for _, pa := range proj.Actions {
				actions = append(actions, Action{
					Name:    pa.Name,
					Command: pa.Command,
				})
			}
		}
	}

	response := ActionsResponse{
		Actions:   actions,
		IsProject: isProject,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
