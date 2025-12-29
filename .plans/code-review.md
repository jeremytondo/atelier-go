# Code Review Implementation Plan

This document outlines the findings from the senior code review and the subsequent plan for implementation.

## 1. Executive Summary
The project is well-structured and follows a standard Go CLI layout (`cmd/` vs `internal/`). The domain logic is reasonably separated (Locations, Sessions, UI). It makes good use of Go's `context` for cancellation and `sync.WaitGroup` for parallelism.

However, there are critical performance risks related to filesystem operations, some unnecessary complexity in path handling, and tight coupling that hampers testability.

## 2. Critical Findings (Performance & Stability)

### 2.1. Severe Performance Bottleneck in Path Canonicalization
**File:** `internal/utils/utils.go` (`GetCanonicalPath`, `correctCasing`)
**Issue:** The application recursively calls `os.ReadDir` for every component of a path to enforce case-sensitivity on case-insensitive filesystems (like macOS).
**Impact:**
*   This function is called for **every** project and **every** `zoxide` result.
*   If `zoxide` returns 50 paths, and each path is 5 levels deep, this could trigger hundreds or thousands of filesystem syscalls (`ReadDir`) just to list locations.
*   This will likely cause noticeable lag on startup.
**Recommendation:** Remove this logic. Standardize on `filepath.EvalSymlinks` and `filepath.Abs`. If case correction is strictly necessary for `zmx` (unlikely).

### 2.2. Inefficient Configuration Loading
**File:** `internal/locations/projects.go`
**Issue:** `viper.New()` is called inside a loop for every TOML file found.
**Impact:** High memory allocation and initialization overhead per file. Viper is a heavy library intended for application-wide config, not high-throughput file parsing.
**Recommendation:** Since the config schema is simple (`Project` struct), use a lighter library like `BurntSushi/toml` or `pelletier/go-toml/v2` directly. If keeping Viper, instantiate it once or reuse the instance.

## 3. Code Quality & Design Review

### 3.1. Testability & Dependency Injection
**File:** `internal/locations/locations.go`
**Issue:** `locations.List` creates concrete providers (`NewProjectProvider`, `NewZoxideProvider`) internally.
**Impact:** It is impossible to unit test `List` or the `Manager` logic without mocking the filesystem or having `zoxide` installed.
**Recommendation:** Refactor `List` to accept a `Manager` or a list of `Provider` interfaces. The CLI layer (`internal/cli`) should be responsible for wiring up the concrete providers.

### 3.2. Fragile UI Input Parsing
**File:** `internal/ui/fzf.go`
**Issue:** The `Select` function attempts to detect keypresses by checking if the output *starts* with the key string.
**Impact:** If a selected item string happens to start with the same characters as a keybinding (e.g., key "a" and item "apple"), the parsing may be ambiguous or incorrect.
**Recommendation:** Use a safe delimiter (like a null byte `\0` or a rare character) in the `fzf` `--bind` argument (e.g., `print(key)+put(\0)+accept`) to strictly separate the metadata from the selection.

### 3.3. Code Duplication
**Files:** `internal/locations/locations.go`, `internal/sessions/sessions.go`
**Issue:** Both files implement nearly identical `PrintTable` logic.
**Recommendation:** Move the table printing logic to `internal/ui` or `internal/utils` as a generic function or a dedicated `TablePrinter` type.

## 4. Nitpicks & Idioms

*   **Error Handling**: In `internal/locations/projects.go`, errors during config loading are silently ignored (`continue`). While this prevents a crash, it leaves the user unaware of broken config files.
    *   *Fix:* Collect errors and return them as a "Warning" list, or print a debug log.
*   **Context Usage**: Excellent. Context is propagated correctly from `main` through to `exec.CommandContext`.
*   **Variable Naming**: Generally idiomatic. `utils.ShortenPath` is clear.
*   **Documentation**: Comments are present and helpful.

## 5. Recommended Action Plan

1.  **Simplify Path Handling (High Priority)**: Delete `utils.correctCasing` and `utils.GetCanonicalPath`. Replace with `filepath.Abs` and `filepath.EvalSymlinks`.
2.  **Refactor Locations**:
    *   Change `locations.List` to `locations.NewManager(...).Fetch(ctx)`.
    *   Move provider instantiation to `internal/cli/locations.go` and `internal/ui/ui.go`.
3.  **Optimize Config Loading**: Replace the `viper` loop in `ProjectProvider` with `pelletier/go-toml/v2` (already an indirect dependency) or `BurntSushi/toml`.
4.  **Consolidate UI**: Create a `ui.PrintTable` function to remove duplication.

## Checklist

- [x] **Simplify Path Handling**
  - [x] Remove `correctCasing` from `internal/utils/utils.go`.
  - [x] Standardize path canonicalization on `filepath.Abs` and `filepath.EvalSymlinks`.
- [x] **Refactor Locations for Dependency Injection**
  - [x] Modify `locations.List` (or `locations.Manager`) to accept providers as arguments rather than instantiating them internally.
  - [x] Update CLI/UI layers to inject the required providers.
- [x] **Optimize Configuration Loading**
  - [x] Centralize configuration loading in `internal/config`.
  - [x] Switch to unified YAML structure (`config.yaml` and `<hostname>.yaml`).
  - [x] Optimize Viper usage (single instance, manual merging).
- [ ] **Consolidate UI Components**
  - [ ] Create a generic `PrintTable` function in `internal/ui` or `internal/utils`.
  - [ ] Refactor `locations` and `sessions` to use the shared table rendering logic.
- [ ] **Harden UI Input Parsing**
  - [ ] Update `Select` in `internal/ui/fzf.go` to use a safe delimiter for key detection instead of `HasPrefix`.

## Implementation Notes

### Path Handling
The `utils.correctCasing` function has been removed and `GetCanonicalPath` has been simplified. We now use `os.Stat` to validate project paths, which properly supports symlinks while avoiding the previous expensive and restrictive recursive canonicalization logic.

### Dependency Injection
We refactored `locations.Manager` to accept a variadic list of `Provider` interfaces in its constructor. The `locations.List` function was removed in favor of this instance-based approach. Provider instantiation and wiring are now centralized in `internal/cli/setup.go` via the `setupLocationManager` helper, and the fully configured manager is passed explicitly to `ui.Run`. This decouples the location logic from specific implementations and significantly improves testability.

### Configuration
Configuration loading has been centralized in `internal/config`. We moved from multiple TOML files to a unified YAML structure (`config.yaml` and `<hostname>.yaml`). Viper usage was optimized by instantiating it once and manually merging projects to avoid slice-replacement issues.

### UI & UX
The `fzf` integration currently relies on string prefix matching which can be ambiguous if project names or paths share common prefixes. A delimited format (e.g., using a null byte or a specific separator) will make result parsing more robust.
