# Specification: Refactor Project Configuration

## 1. Overview
This track focuses on refactoring the existing `internal/config` package to align with the product definition and tech stack. The goal is to robustly implement YAML-based configuration using `spf13/viper`, ensuring clear separation of concerns, support for host-specific overrides, and a strongly typed configuration structure.

## 2. Goals
- **Clean Architecture:** Refactor `internal/config` to be modular and testable.
- **Viper Integration:** Fully leverage `spf13/viper` for loading, parsing, and watching configuration files.
- **Host-Specific Overrides:** Implement logic to merge a base `config.yaml` with a host-specific file (e.g., `<hostname>.yaml`).
- **Strong Typing:** Define Go structs that map directly to the configuration schema, ensuring type safety.
- **Validation:** Add basic validation to ensure essential configuration fields are present and correct.

## 3. User Stories
- As a developer, I want to define my project settings in a clean `config.yaml` file so I can easily version control my dotfiles.
- As a user with multiple machines, I want to have a base configuration and a host-specific override so I can handle path differences without duplicating the entire config.
- As a maintainer, I want the configuration code to be isolated and well-tested so that future changes don't break the application initialization.

## 4. Technical Requirements
- **Package:** `internal/config`
- **Dependencies:** `github.com/spf13/viper`
- **Structs:**
    - `Config` (Root)
    - `Theme` (UI Colors)
    - `Project` (Project definition)
    - `Action` (Command definition)
- **Logic:**
    1.  Initialize Viper.
    2.  Set defaults.
    3.  Load `config.yaml` from `~/.config/atelier-go/`.
    4.  Determine hostname.
    5.  Attempt to load `<hostname>.yaml` and merge it.
    6.  Unmarshal into the `Config` struct.
    7.  Validate the struct.

## 5. Non-Functional Requirements
- **Error Handling:** Configuration errors should return clear, actionable messages to the user (e.g., "Malformed YAML in config.yaml").
- **Performance:** Configuration loading should be negligible in startup time.
- **Testing:** Unit tests should cover default loading, merging logic, and validation failures.
