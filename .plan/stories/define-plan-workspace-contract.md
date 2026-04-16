---
created_at: "2026-04-16T05:46:33Z"
epic: core-workspace-and-artifact-system
project: plan
slug: define-plan-workspace-contract
spec: core-workspace-and-artifact-system
status: todo
title: Define .plan workspace contract
type: story
updated_at: "2026-04-16T05:46:33Z"
---

# Define .plan workspace contract

Created: 2026-04-16T05:46:33Z

## Description

Finalize the v1 `.plan/` layout and root artifact contract so the workspace model is stable and easy to explain.
## Acceptance Criteria


- [ ] Workspace layout is explicitly defined for PROJECT.md, ROADMAP.md, epics, specs, stories, and .meta

- [ ] The contract stays local-first and markdown-first

- [ ] The scope boundary between user-authored notes and tool-owned state is clear
## Verification


- go test ./internal/workspace ./cmd
## Resources


- [Canonical Spec](../specs/core-workspace-and-artifact-system.md)
## Notes
