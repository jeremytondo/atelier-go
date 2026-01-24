package config

import (
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Valid Config",
			config: Config{
				Projects: []Project{
					{Name: "p1", Path: "/tmp"},
				},
				Editor: "vim",
			},
			wantErr: false,
		},
		{
			name: "Project Missing Path",
			config: Config{
				Projects: []Project{
					{Name: "p1", Path: ""},
				},
			},
			wantErr: true,
		},
		{
			name: "Project Missing Name",
			config: Config{
				Projects: []Project{
					{Name: "", Path: "/tmp"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
