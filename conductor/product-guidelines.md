# Product Guidelines: Atelier Go

## Documentation & Voice
- **Approachable and Educational:** Documentation and UI text should be helpful and guiding. We aim to explain the "why" behind features, providing context that helps users integrate Atelier Go into their native terminal workflow smoothly.

## Visual & UI Design
- **Modern & Vibrant:** The TUI should feel like a modern application. We leverage `lipgloss` for rich colors, utilizing borders and styles that make the interface clear and visually engaging while respecting terminal constraints.
- **Visual Hierarchy:** Use colors and styles to clearly distinguish between project names, paths, and available actions. Focus should be unmistakable.

## Error Handling & Feedback
- **Graceful & Informative:** Errors should be caught and presented to the user as actionable information within the UI. Avoid crashing or dumping raw stack traces. The goal is to keep the user informed and in control of their session.

## Performance & Responsiveness
- **Instantaneous Feedback:** As a launcher, speed is a feature. Interaction—especially fuzzy searching and navigation—must feel immediate. We prioritize asynchronous processing where necessary to ensure the UI remains fluid.

## Keyboard Navigation & Interaction
- **Discoverable & Efficient:** We provide a dual approach to navigation:
    - **Standardized:** Support for common keys like arrows, Enter, and Escape for immediate familiarity.
    - **Power User Focused:** Full support for Vim-like bindings (`h`, `j`, `k`, `l`, etc.) for high-efficiency navigation.
- **Visible Help:** A persistent help menu or key should always be available to remind users of available shortcuts.
