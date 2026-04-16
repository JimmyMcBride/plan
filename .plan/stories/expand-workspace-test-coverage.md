---
created_at: "2026-04-16T05:46:33Z"
epic: core-workspace-and-artifact-system
project: plan
slug: expand-workspace-test-coverage
spec: core-workspace-and-artifact-system
status: done
title: Expand workspace test coverage
type: story
updated_at: "2026-04-16T06:06:39Z"
---

# Expand workspace test coverage

Created: 2026-04-16T05:46:33Z

## Description

Add focused tests around initialization, repair, and upgrade-safe behavior for the `.plan/` workspace.
## Acceptance Criteria


- [ ] Init coverage includes full v1 workspace layout

- [ ] Repair coverage includes partial workspace states

- [ ] Regression coverage protects the workspace contract from accidental drift
## Verification


- go test ./internal/workspace ./...
## Resources


- [Canonical Spec](../specs/core-workspace-and-artifact-system.md)
## Notes
