# Implementation Plan - Refactor Project Configuration

This plan outlines the steps to refactor the configuration package to support modular YAML configuration and host-specific overrides using Viper.

## Phase 1: Struct Definition & Defaults
- [x] Task: Define the `Config` struct and its children (`Theme`, `Project`, `Action`) in `internal/config/types.go` with appropriate mapstructure tags. c6052e6
    - [x] Subtask: Create `types.go` and define structs.
    - [x] Subtask: Write unit tests to verify struct tags and basic instantiation.
- [~] Task: Implement default configuration values.
    - [ ] Subtask: Create a `SetDefaults` function using Viper.
    - [ ] Subtask: Write tests to ensure defaults are populated when no config file is present.
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Struct Definition & Defaults' (Protocol in workflow.md)

## Phase 2: Loading & Merging Logic
- [ ] Task: Implement base configuration loading.
    - [ ] Subtask: Create `LoadConfig` function in `internal/config/config.go`.
    - [ ] Subtask: Configure Viper to look in `~/.config/atelier-go` and read `config.yaml`.
    - [ ] Subtask: Write tests for successful and failed file loading.
- [ ] Task: Implement host-specific override logic.
    - [ ] Subtask: Implement logic to detect hostname.
    - [ ] Subtask: Use Viper's `MergeInConfig` to overlay `<hostname>.yaml`.
    - [ ] Subtask: Create integration tests with mock config files to verify merging behavior (e.g., host config overriding base config).
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Loading & Merging Logic' (Protocol in workflow.md)

## Phase 3: Validation & Cleanup
- [ ] Task: Implement configuration validation.
    - [ ] Subtask: Add a `Validate` method to the `Config` struct.
    - [ ] Subtask: Ensure critical fields (e.g., project paths) are checked.
    - [ ] Subtask: Write tests for validation edge cases.
- [ ] Task: Refactor existing code to use the new `config` package.
    - [ ] Subtask: update `cmd/atelier-go/main.go` to use the new loader.
    - [ ] Subtask: Update `internal/ui` or other consumers to use the new `Config` struct.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Validation & Cleanup' (Protocol in workflow.md)
