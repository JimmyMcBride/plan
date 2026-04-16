---
created_at: "2026-04-16T06:28:47Z"
epic: brain-interop-and-planning-imports
project: plan
slug: import-brain-planning-notes-into-plan-artifacts
spec: brain-interop-and-planning-imports
status: done
title: Import brain planning notes into plan artifacts
type: story
updated_at: "2026-04-16T07:07:28Z"
---

# Import brain planning notes into plan artifacts

Created: 2026-04-16T06:28:47Z

## Description

Map selected brain planning material into plan epics, specs, and stories.
## Acceptance Criteria

- [ ] Imported notes land under .plan using canonical plan metadata and file locations.

- [ ] Import keeps plan responsible only for planning artifacts after creation.
## Verification

- go test ./internal/planning
## Resources

- [Canonical Spec](../specs/brain-interop-and-planning-imports.md)
## Notes
