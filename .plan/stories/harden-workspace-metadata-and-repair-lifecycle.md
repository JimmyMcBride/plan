---
created_at: "2026-04-16T05:46:33Z"
epic: core-workspace-and-artifact-system
project: plan
slug: harden-workspace-metadata-and-repair-lifecycle
spec: core-workspace-and-artifact-system
status: done
title: Harden workspace metadata and repair lifecycle
type: story
updated_at: "2026-04-16T06:05:34Z"
---

# Harden workspace metadata and repair lifecycle

Created: 2026-04-16T05:46:33Z

## Description


Strengthen workspace metadata, doctor output, and update behavior so partial or stale workspaces can be understood and repaired safely.
## Acceptance Criteria


- [ ] Doctor reports current, missing, or broken state clearly

- [ ] Update repairs tool-managed surfaces without damaging user-authored planning notes

- [ ] Workspace metadata and migration state remain small and inspectable
## Verification


- go test ./internal/workspace
## Resources


- [Canonical Spec](../specs/core-workspace-and-artifact-system.md)
## Notes
