---
created_at: "2026-04-16T06:28:47Z"
epic: workspace-adoption-update-and-migration
project: plan
slug: add-adopt-command-for-existing-repos
spec: workspace-adoption-update-and-migration
status: todo
title: Add adopt command for existing repos
type: story
updated_at: "2026-04-16T06:28:47Z"
---

# Add adopt command for existing repos

Created: 2026-04-16T06:28:47Z

## Description

Create an explicit local-first adoption flow for repos that do not yet use plan.
## Acceptance Criteria

- [ ] Adopt creates only plan-managed surfaces and leaves non-plan repo files untouched.

- [ ] Adopt produces a current workspace that doctor and update can manage afterward.
## Verification

- go test ./cmd ./internal/workspace
## Resources

- [Canonical Spec](../specs/workspace-adoption-update-and-migration.md)
## Notes
