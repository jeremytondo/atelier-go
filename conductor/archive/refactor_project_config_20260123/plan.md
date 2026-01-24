# Implementation Plan - Refactor Project Configuration

This plan outlines the steps to refactor the configuration package to support modular YAML configuration and host-specific overrides using Viper.

## Phase 1: Struct Definition & Defaults [checkpoint: bf695f3]
- [x] Task: Define the `Config` struct and its children (`Theme`, `Project`, `Action`) in `internal/config/types.go` with appropriate mapstructure tags. c6052e6
    - [x] Subtask: Create `types.go` and define structs.
    - [x] Subtask: Write unit tests to verify struct tags and basic instantiation.
- [x] Task: Implement default configuration values. cdf31d4
    - [x] Subtask: Create a `SetDefaults` function using Viper.
    - [x] Subtask: Write tests to ensure defaults are populated when no config file is present.
- [x] Task: Conductor - User Manual Verification 'Phase 1: Struct Definition & Defaults' (Protocol in workflow.md)

## Phase 2: Loading & Merging Logic
- [x] Task: Implement base configuration loading. 4c8abf2
    - [x] Subtask: Create `LoadConfig` function in `internal/config/config.go`.
    - [x] Subtask: Configure Viper to look in `~/.config/atelier-go` and read `config.yaml`.
    - [x] Subtask: Write tests for successful and failed file loading.
- [x] Task: Implement host-specific override logic. 4c8abf2
    - [x] Subtask: Implement logic to detect hostname.
    - [x] Subtask: Use Viper's `MergeInConfig` to overlay `<hostname>.yaml`.
    - [x] Subtask: Create integration tests with mock config files to verify merging behavior (e.g., host config overriding base config).
- [~] Task: Conductor - User Manual Verification 'Phase 2: Loading & Merging Logic' (Protocol in workflow.md)

## Phase 3: Validation & Cleanup [checkpoint: bd83b02]
- [x] Task: Implement configuration validation. 90203a7
    - [x] Subtask: Add a `Validate` method to the `Config` struct.
    - [x] Subtask: Ensure critical fields (e.g., project paths) are checked.
    - [x] Subtask: Write tests for validation edge cases.
- [x] Task: Refactor existing code to use the new `config` package. 90203a7
    - [x] Subtask: update `cmd/atelier-go/main.go` to use the new loader.
    - [x] Subtask: Update `internal/ui` or other consumers to use the new `Config` struct.
- [x] Task: Conductor - User Manual Verification 'Phase 3: Validation & Cleanup' (Protocol in workflow.md)
