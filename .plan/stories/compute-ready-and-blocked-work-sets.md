---
created_at: "2026-04-16T06:28:47Z"
epic: dependency-graph-and-ready-work
project: plan
slug: compute-ready-and-blocked-work-sets
spec: dependency-graph-and-ready-work
status: todo
title: Compute ready and blocked work sets
type: story
updated_at: "2026-04-16T06:28:47Z"
---

# Compute ready and blocked work sets

Created: 2026-04-16T06:28:47Z

## Description

Derive which stories are ready now and which are blocked by unfinished dependencies.
## Acceptance Criteria

- [ ] Ready/blocked evaluation explains which blockers keep a story from moving.

- [ ] Dependency evaluation respects current story lifecycle states and completed work.
## Verification

- go test ./internal/planning
## Resources

- [Canonical Spec](../specs/dependency-graph-and-ready-work.md)
## Notes
