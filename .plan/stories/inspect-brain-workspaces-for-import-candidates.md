---
created_at: "2026-04-16T06:28:47Z"
epic: brain-interop-and-planning-imports
project: plan
slug: inspect-brain-workspaces-for-import-candidates
spec: brain-interop-and-planning-imports
status: todo
title: Inspect brain workspaces for import candidates
type: story
updated_at: "2026-04-16T06:28:47Z"
---

# Inspect brain workspaces for import candidates

Created: 2026-04-16T06:28:47Z

## Description

Add a read-only inspection flow for planning material that already exists in brain workspaces.
## Acceptance Criteria

- [ ] Plan can detect importable planning notes from a local brain workspace without changing files.

- [ ] Preview output stays focused on planning artifacts instead of brain memory or session data.
## Verification

- go test ./internal/planning
## Resources

- [Canonical Spec](../specs/brain-interop-and-planning-imports.md)
## Notes
