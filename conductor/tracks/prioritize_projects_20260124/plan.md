# Implementation Plan: Prioritize Projects in Locations List

## Phase 1: Logic & Sorting Implementation
- [x] Task: Update the internal location sorting logic to support multi-level sorting (Type then Score). 374479d
- [x] Task: Write Tests: Verify that projects are always sorted above directories in various scenarios (empty search, fuzzy match). e19c0ab
- [ ] Task: Implement: Modify the filtering/sorting function in `internal/ui/filter.go` (or equivalent) to enforce project priority.
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Logic & Sorting Implementation' (Protocol in workflow.md)

## Phase 2: UI Integration & Verification
- [ ] Task: Ensure the UI model correctly handles the prioritized list without visual regressions.
- [ ] Task: Write Tests: Verify the UI components render the sorted list as expected.
- [ ] Task: Implement: Any necessary adjustments to `internal/ui/model.go` or `internal/ui/view.go` to support the prioritized list.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: UI Integration & Verification' (Protocol in workflow.md)
