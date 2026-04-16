---
created_at: "2026-04-16T06:28:47Z"
epic: workspace-adoption-update-and-migration
project: plan
slug: expand-migration-tracking-and-repair-coverage
spec: workspace-adoption-update-and-migration
status: todo
title: Expand migration tracking and repair coverage
type: story
updated_at: "2026-04-16T06:28:47Z"
---

# Expand migration tracking and repair coverage

Created: 2026-04-16T06:28:47Z

## Description

Strengthen migration state so repeated repair and upgrade flows remain inspectable over time.
## Acceptance Criteria

- [ ] Migration tracking records enough detail to explain what plan repaired or normalized.

- [ ] Tests cover repeated adopt, doctor, and update flows without regressing idempotency.
## Verification

- go test ./internal/workspace ./cmd
## Resources

- [Canonical Spec](../specs/workspace-adoption-update-and-migration.md)
## Notes
