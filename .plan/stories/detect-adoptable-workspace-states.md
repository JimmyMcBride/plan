---
created_at: "2026-04-16T06:28:47Z"
epic: workspace-adoption-update-and-migration
project: plan
slug: detect-adoptable-workspace-states
spec: workspace-adoption-update-and-migration
status: todo
title: Detect adoptable workspace states
type: story
updated_at: "2026-04-16T06:28:47Z"
---

# Detect adoptable workspace states

Created: 2026-04-16T06:28:47Z

## Description

Teach plan to distinguish a missing workspace from an existing repo that is safe to adopt.
## Acceptance Criteria

- [ ] Doctor reports adoptable, partial, missing, and broken states with clear repair guidance.

- [ ] Detection stays limited to plan-managed surfaces and does not inspect unrelated repo files aggressively.
## Verification

- go test ./internal/workspace
## Resources

- [Canonical Spec](../specs/workspace-adoption-update-and-migration.md)
## Notes
