package config

// Project represents a defined project with a name and a filesystem path.
type Project struct {
	Name           string   `mapstructure:"name"`
	Path           string   `mapstructure:"path"`
	Actions        []Action `mapstructure:"actions"`
	DefaultActions *bool    `mapstructure:"default-actions"`
	ShellDefault   *bool    `mapstructure:"shell-default"`
}

// Action represents a runnable command associated with a project.
type Action struct {
	Name    string `mapstructure:"name"`
	Command string `mapstructure:"command"`
}

// Config represents the application configuration.
type Config struct {
	Projects     []Project `mapstructure:"projects"`
	Actions      []Action  `mapstructure:"actions"`
	ShellDefault *bool     `mapstructure:"shell-default"`
	Editor       string    `mapstructure:"editor"`
	Theme        Theme     `mapstructure:"theme"`
}

// Theme holds color settings for the UI.
type Theme struct {
	Primary   string `mapstructure:"primary"`
	Accent    string `mapstructure:"accent"`
	Highlight string `mapstructure:"highlight"`
	Text      string `mapstructure:"text"`
	Subtext   string `mapstructure:"subtext"`
}
