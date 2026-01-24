# Specification: Prioritize Projects in Locations List

## Overview
Currently, the locations list (projects and zoxide directories) is sorted purely by fuzzy match score or usage frequency. This track aims to prioritize projects by always displaying them at the top of the list, regardless of whether a search query is active or empty.

## Functional Requirements
- **Strict Grouping:** In the UI list, all matching "Projects" must appear before any matching "Directories" (zoxide paths).
- **Persistent Priority:** This sorting rule applies both to the initial view (empty search) and while fuzzy searching.
- **Visual Consistency:** The list will remain flat without explicit section headers, maintaining the existing UI aesthetic.

## Non-Functional Requirements
- **Performance:** Sorting logic should not introduce perceptible lag during fuzzy searching.

## Acceptance Criteria
- [ ] When opening the UI, all configured projects are listed before any zoxide directories.
- [ ] When typing a search query, any project matching the query appears above any zoxide directory matching the query.
- [ ] Fuzzy matching quality still determines the relative order *within* the project group and *within* the directory group.

## Out of Scope
- Adding UI section headers or categories.
- Changing the underlying fuzzy matching algorithm.
