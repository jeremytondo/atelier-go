---
status: completed
updated: 2026-01-03
---

# Default Actions

## Idea
It would be nice if there was a way to have a set of default actions defined in 
the config. These actions would show up for all locations, even for projects.

We could define these in an actions collection at the top level of the config.
They could be defined in the global config or host specific config and follow the
same merging rules we've already set. It could look something like this:

```yaml
actions:
  - name: "Action Name"
    command: "action commmand"
```

For projects we could add a setting to maybe turn off the default actions if
a user wanted to. We could add an option like:

default-actions: true

If this option is not in a project config, it will default to true. If it exists
and is set to false, the default actions would not show up for that project.

## Plan

### Key Clarifications

1. **All locations receive default actions** - Not just projects. Zoxide locations (and any future non-project providers) will also receive default actions. This is a change from current behavior where non-project locations only have a single hardcoded "shell" action.

2. **Config merging follows existing patterns** - Actions should be merged the same way Projects are merged: by name, where host-specific actions override global actions if names match, otherwise append. Add a `mergeActions` function that mirrors `mergeProjects`.

3. **No unit tests** - Tests are not required for this implementation.

4. **Replaces hardcoded "shell" action** - The current hardcoded "shell" action for non-project locations will be replaced by the configurable default actions system.

5. **Shell is always available as a built-in action** - "Shell" is not a configured action but a built-in behavior. It appears in the UI action list whenever there are other configured actions, appended at the end. If there are NO configured actions (no default actions and no project-specific actions), the action panel is not shown and selecting the location opens a shell directly.

6. **`shell-default` controls Shell action position** - A new `shell-default` configuration option (default: `false`) controls whether Shell is the default action:
   - `shell-default: false` (default): Shell is appended to the end of the action list. The first configured action becomes the default.
   - `shell-default: true`: Shell is prepended to the beginning of the action list, making it the default.
   - Projects inherit the root `shell-default` value but can override it.
   - If no actions are configured, behavior is unchanged (opens shell directly, no action panel).

### 1. Configuration Updates (`internal/config`)
- **File**: `internal/config/config.go`
- **Changes**:
    - Add `Actions []Action` to the `Config` struct.
    - Add `mergeActions` helper function (mirrors `mergeProjects` logic).
    - Update `LoadConfig` to merge `Actions` using `mergeActions`.
    - Add `DefaultActions *bool` to the `Project` struct (YAML: `default-actions`).
    - Implement a helper `(p Project) UseDefaultActions() bool` that defaults to `true`.

### 2. Location Provider Updates (`internal/locations`)
- **File**: `internal/locations/projects.go`
- **Changes**:
    - Update `ProjectProvider` constructor to accept default actions.
    - Update `ProjectProvider` to store `defaultActions []config.Action`.
    - In `Fetch`, if `UseDefaultActions()` is true, merge `defaultActions` with the `Location.Actions` (respecting `UseDefaultActions`).
- **File**: `internal/locations/zoxide.go`
- **Changes**:
    - Update `ZoxideProvider` constructor to accept default actions.
    - Update `ZoxideProvider` to store `defaultActions []config.Action`.
    - In `Fetch`, include `defaultActions` on all `Location` objects.

### 3. CLI Integration
- **File**: `internal/cli/setup.go`
- **Changes**:
    - Update `setupLocationManager` to pass `cfg.Actions` to providers.

### 4. Session Resolution Updates (`internal/sessions`)
- **File**: `internal/sessions/sessions.go`
- **Changes**:
    - Refactor `Resolve` function to remove the `loc.Source == "Project"` branching
    - Use unified action-based logic for all location types:
        1. If an action name is provided (and not "shell"), look it up in `loc.Actions`
        2. If "shell" is explicitly requested OR no actions exist, open a shell
        3. If no action name provided and actions exist, use the first action as default
        4. Keep "editor" as a special case for backwards compatibility
    - This simplifies the code by treating all locations uniformly once they have actions

### 5. UI Updates (`internal/ui`)
- **File**: `internal/ui/model.go`
- **Changes**:
    - Update `updateActions()` to add "Shell" option for any location that has actions (not just Projects)
    - Remove the `sel.IsProject()` check
    - "Shell" is a built-in action that always appears at the end of the action list when there are configured actions
    - If a location has no actions, the action panel remains empty and selecting the location opens a shell directly (existing behavior preserved)

### 6. Shell-Default Configuration
- **File**: `internal/config/config.go`
- **Changes**:
    - Add `ShellDefault *bool` field to `Config` struct (YAML: `shell-default`)
    - Add `ShellDefault *bool` field to `Project` struct (YAML: `shell-default`)
    - Add `GetShellDefault() bool` helper method on `Config` (returns `false` if nil)
    - Add `GetShellDefault(rootDefault bool) bool` helper method on `Project` (inherits from root if nil)

- **File**: `internal/locations/projects.go`
- **Changes**:
    - Update `ProjectProvider` to store `rootShellDefault bool`
    - Update constructor to accept `rootShellDefault`
    - Create `buildActionsWithShell()` helper to position Shell correctly based on `shellDefault`
    - Update `Fetch` to use the helper and determine effective `shellDefault` per project

- **File**: `internal/locations/zoxide.go`
- **Changes**:
    - Update `ZoxideProvider` to store `shellDefault bool`
    - Update constructor to accept `shellDefault`
    - Update `Fetch` to use `buildActionsWithShell()` helper

- **File**: `internal/cli/setup.go`
- **Changes**:
    - Pass `cfg.GetShellDefault()` to both providers

- **File**: `internal/ui/model.go`
- **Changes**:
    - Remove the code that appends "Shell" in `updateActions()` (Shell is now added by providers)

### 7. Logic for Merging Actions
- When merging project-specific actions and default actions, project-specific actions should come first.
- If an action with the same name exists in both, the project-specific one should take precedence (effectively overriding the default).

## Execution

- [x] Add `Actions []Action` field to `config.Config` struct
- [x] Add `mergeActions` helper function (mirrors `mergeProjects` logic)
- [x] Update `LoadConfig` to merge `Actions` using `mergeActions`
- [x] Add `DefaultActions *bool` field to `config.Project` struct (YAML: `default-actions`)
- [x] Add `UseDefaultActions() bool` helper method on `Project`
- [x] Update `ProjectProvider` constructor to accept default actions
- [x] Update `ProjectProvider.Fetch` to merge default actions (respecting `UseDefaultActions`)
- [x] Update `ZoxideProvider` constructor to accept default actions
- [x] Update `ZoxideProvider.Fetch` to include default actions on all locations
- [x] Update `setupLocationManager` in `internal/cli/setup.go` to pass `cfg.Actions` to providers
- [x] Refactor `sessions.Resolve` to use unified action-based logic (remove `loc.Source == "Project"` branching)
- [x] Update `ui/model.go` to show "Shell" option for all locations with actions
- [x] Add `ShellDefault *bool` field to `config.Config` struct
- [x] Add `ShellDefault *bool` field to `config.Project` struct
- [x] Add `GetShellDefault()` helper method on `Config`
- [x] Add `GetShellDefault(rootDefault bool)` helper method on `Project`
- [x] Create `buildActionsWithShell()` helper function in locations package
- [x] Update `ProjectProvider` constructor to accept `shellDefault`
- [x] Update `ProjectProvider.Fetch` to use `buildActionsWithShell()`
- [x] Update `ZoxideProvider` constructor to accept `shellDefault`
- [x] Update `ZoxideProvider.Fetch` to use `buildActionsWithShell()`
- [x] Update `setupLocationManager` to pass shell-default to providers
- [x] Remove Shell appending logic from `ui/model.go`

## Implementation Summary

The default actions feature has been fully implemented, allowing users to define global actions that apply to all locations. Key highlights include:
- **Global Actions**: Defined in the root config and merged across host-specific configurations.
- **Project Overrides**: Projects can opt-out of default actions using `default-actions: false`.
- **Shell Action Positioning**: A new `shell-default` setting (at both root and project levels) controls whether the built-in "Shell" action appears at the beginning or end of the action list.
- **Unified Logic**: Both Project and Zoxide providers now share a common action-building logic via `buildActionsWithShell`, and `sessions.Resolve` has been simplified to handle all location types uniformly.
- **UI Decoupling**: The UI no longer manages the "Shell" action injection; it simply displays the actions provided by the location manager, which are now pre-sorted and include the shell action in the desired position.

