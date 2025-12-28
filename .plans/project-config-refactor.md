# Project Configuration Refactor

## Goal
Transition from a single inline `projects` list in `config.toml` to individual `.toml` files stored in a `projects/` subdirectory of the configuration directory (respecting XDG conventions).

## Technical Requirements

### 1. Configuration Path Management
- **File**: `internal/config/config.go`
- **Changes**:
    - Export `GetConfigDir() (string, error)` to handle `$XDG_CONFIG_HOME/atelier-go` or `~/.config/atelier-go`.
    - Remove `Projects []Project` from the main `Config` struct.

### 2. Path Utilities
- **File**: `internal/utils/utils.go`
- **Changes**:
    - Implement `ExpandPath(path string) (string, error)` to support `~` expansion and environment variable resolution.

### 3. Project Provider Refactoring
- **File**: `internal/locations/projects.go`
- **Changes**:
    - Update `ProjectProvider` to load projects exclusively from files in `$CONFIG_DIR/projects/`.
    - Create the `projects/` directory if it doesn't exist.
    - Each `.toml` file unmarshals into the existing `config.Project` struct.
    - Process all `path` fields through `ExpandPath`.

### 4. CLI Integration
- **File**: `internal/cli/cli.go`
- **Changes**:
    - Update to pass the config directory to the provider.

## Example Project File (`projects/my-app.toml`)
```toml
name = "My App"
path = "~/code/my-app"

[[actions]]
name = "test"
command = "go test ./..."
```

## Implementation Notes
- Use `os.UserConfigDir()` as a starting point for XDG compliance.
- The `projects/` directory should be automatically created if missing to ensure a smooth user experience.
- Each file in `projects/` should be a standalone TOML file representing a single project.
- Error handling should be robust, especially when reading and unmarshaling multiple files.

## Checklist
- [ ] Export `GetConfigDir` in `internal/config/config.go`
- [ ] Remove `Projects` field from `config.Config` struct
- [ ] Implement `ExpandPath` in `internal/utils/utils.go`
- [ ] Refactor `ProjectProvider` in `internal/locations/projects.go` to load from directory
- [ ] Ensure `projects/` directory is created if missing
- [ ] Apply `ExpandPath` to all project paths
- [ ] Update `internal/cli/cli.go` to initialize provider with config path
- [ ] Verify local projects are correctly discovered from the new directory structure
