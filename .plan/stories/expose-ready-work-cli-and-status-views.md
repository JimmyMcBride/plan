---
created_at: "2026-04-16T06:28:47Z"
epic: dependency-graph-and-ready-work
project: plan
slug: expose-ready-work-cli-and-status-views
spec: dependency-graph-and-ready-work
status: todo
title: Expose ready-work CLI and status views
type: story
updated_at: "2026-04-16T06:28:47Z"
---

# Expose ready-work CLI and status views

Created: 2026-04-16T06:28:47Z

## Description

Add a user-facing view for dependency-aware work selection.
## Acceptance Criteria

- [ ] CLI output can show ready stories and blocked stories with blocking reasons.

- [ ] Status views can surface dependency-aware progress without replacing the base story model.
## Verification

- go test ./cmd ./internal/planning
## Resources

- [Canonical Spec](../specs/dependency-graph-and-ready-work.md)
## Notes
