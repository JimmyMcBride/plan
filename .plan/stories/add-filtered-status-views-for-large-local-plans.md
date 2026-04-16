---
created_at: "2026-04-16T06:28:47Z"
epic: power-user-local-workflows
project: plan
slug: add-filtered-status-views-for-large-local-plans
spec: power-user-local-workflows
status: done
title: Add filtered status views for large local plans
type: story
updated_at: "2026-04-16T07:12:42Z"
---

# Add filtered status views for large local plans

Created: 2026-04-16T06:28:47Z

## Description

Improve status output so larger local plans stay workable without abandoning the simple model.
## Acceptance Criteria

- [ ] Status commands can filter or narrow output by version, epic, or lifecycle slice.

- [ ] Filtered views still keep the default status path clean for small repos.
## Verification

- go test ./cmd ./internal/planning
## Resources

- [Canonical Spec](../specs/power-user-local-workflows.md)
## Notes
