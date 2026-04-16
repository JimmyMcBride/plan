---
created_at: "2026-04-16T06:28:47Z"
epic: power-user-local-workflows
project: plan
slug: add-advanced-roadmap-and-story-selection-helpers
spec: power-user-local-workflows
status: done
title: Add advanced roadmap and story selection helpers
type: story
updated_at: "2026-04-16T07:15:10Z"
---

# Add advanced roadmap and story selection helpers

Created: 2026-04-16T06:28:47Z

## Description

Add helpers that make next-step selection easier on larger repos with many active stories.
## Acceptance Criteria

- [ ] CLI helpers can narrow work by version, epic, or story state using local plan data.

- [ ] Selection helpers remain local-first and grounded in markdown-backed state rather than hidden queues.
## Verification

- go test ./cmd ./internal/planning
## Resources

- [Canonical Spec](../specs/power-user-local-workflows.md)
## Notes
