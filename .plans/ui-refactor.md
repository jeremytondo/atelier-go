# UI Refactor: Native TUI with Bubble Tea

## Goal
Replace the existing `fzf` dependency with a native, elegant Go TUI built using the Charm ecosystem. This will provide a more integrated experience, better performance, and a highly customizable "Spotlight-style" interface.

## Tech Stack
- **Bubble Tea**: The TUI framework (Model-Update-View).
- **Bubbles**: Standard components (list, textinput, spinner).
- **Lip Gloss**: Styling and layout definitions.

## Core UX (Spotlight Mode)
- **Minimal Start**: The application launches as a minimal, centered search box with a search icon (``).
- **Dynamic Expansion**: The UI dynamically expands vertically to show results as the user begins typing.
- **Master-Detail Layout**: 
    - **Left Panel (Master)**: Displays Locations (Projects/Folders).
    - **Right Panel (Detail)**: Displays context-aware Actions for the currently selected location (e.g., "Attach", "Open in Editor", "New Session").

## Navigation & Interaction
- **Universal Search**: Filtering is focused on the active panel.
- **Master to Detail**: Press `Enter` or `Tab` to move focus from the Locations list to the Actions list.
- **Fast Select**: Press `Alt+Enter` or `Ctrl+S` to execute the default action immediately without entering the detail view.
- **Back/Cancel**: `Esc` navigates back from detail to master, or cancels/exits if already at the master level.
- **List Navigation**: Supports `Ctrl+N`/`Ctrl+P` and Arrow Keys.

## Visual Style
- **Layout**: Centered placement on the screen with rounded borders.
- **Iconography**:
    - Project icons: ``
    - Folder icons: `` (uea83)
- **Theming**:
    - Pink/Purple selection highlights for the active element.
    - Dimmed/Greyed-out state for unfocused panels to maintain visual hierarchy.

## Implementation Strategy
- **Abstraction**: Refactor `internal/ui` to abstract the selection logic away from `fzf` specific implementations.
- **Architecture**: Implement the standard Bubble Tea `Model`, `Update`, and `View` functions.
- **Display**: Use the Alternate Screen buffer to ensure a clean exit and restoration of the terminal state.
- **Responsiveness**: Handle terminal resize events to keep the Spotlight interface centered.

## Checklist
- [ ] Research Bubble Tea list filtering performance for large project counts.
- [ ] Define the `Model` state for Master-Detail navigation.
- [ ] Create Lip Gloss styles for the search box and panels.
- [ ] Implement the `Locations` provider integration into the TUI.
- [ ] Implement the `Actions` provider integration.
- [ ] Add keybindings for Fast Select and navigation.
- [ ] Refactor `internal/cli` to use the new TUI instead of the `fzf` runner.
- [ ] Add support for icons (Nerd Fonts).

## Implementation Notes
- The TUI should gracefully handle cases where Nerd Fonts are not available (fallback to simple characters).
- Ensure that the search input is always focused when the user starts typing, regardless of which panel is selected (unless explicitly navigating the list).
- Consider using `bubbles/list` for the panels but customize the delegate for the Master-Detail look.
