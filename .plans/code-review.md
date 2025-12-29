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

- [ ] **Simplify Path Handling**
  - [ ] Remove `correctCasing` from `internal/utils/utils.go`.
  - [ ] Standardize path canonicalization on `filepath.Abs` and `filepath.EvalSymlinks`.
- [ ] **Refactor Locations for Dependency Injection**
  - [ ] Modify `locations.List` (or `locations.Manager`) to accept providers as arguments rather than instantiating them internally.
  - [ ] Update CLI/UI layers to inject the required providers.
- [ ] **Optimize Configuration Loading**
  - [ ] Remove Viper dependency for project configuration.
  - [ ] Implement configuration parsing using `pelletier/go-toml/v2`.
  - [ ] Ensure `ProjectProvider` does not instantiate parsers in a loop.
- [ ] **Consolidate UI Components**
  - [ ] Create a generic `PrintTable` function in `internal/ui` or `internal/utils`.
  - [ ] Refactor `locations` and `sessions` to use the shared table rendering logic.
- [ ] **Harden UI Input Parsing**
  - [ ] Update `Select` in `internal/ui/fzf.go` to use a safe delimiter for key detection instead of `HasPrefix`.

## Implementation Notes

### Path Handling
The current `utils.correctCasing` implementation uses recursive `os.ReadDir`, which causes significant performance degradation. Standard Go library functions `filepath.Abs` and `filepath.EvalSymlinks` should be used to handle path normalization more efficiently.

### Dependency Injection
Hardcoded providers in `internal/locations` make unit testing difficult. Moving provider instantiation to the entry point (CLI) and passing them down via interfaces will improve testability.

### Configuration
Viper's overhead is unnecessary for the project's needs, especially when called within loops. Switching to `go-toml/v2` provides a lighter, faster alternative for TOML parsing.

### UI & UX
The `fzf` integration currently relies on string prefix matching which can be ambiguous if project names or paths share common prefixes. A delimited format (e.g., using a null byte or a specific separator) will make result parsing more robust.
