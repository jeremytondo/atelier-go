package api

import (
	"encoding/json"
	"net/http"

	"github.com/spf13/viper"
)

type Action struct {
	Name    string `json:"name" mapstructure:"name"`
	Command string `json:"command" mapstructure:"command"`
}

type ActionsResponse struct {
	Actions []Action `json:"actions"`
}

func ActionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var actions []Action
	if err := viper.UnmarshalKey("actions", &actions); err != nil {
		// If unmarshal fails, we start with an empty list
		actions = []Action{}
	}

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

	response := ActionsResponse{
		Actions: actions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
