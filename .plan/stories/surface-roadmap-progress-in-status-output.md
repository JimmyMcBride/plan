---
created_at: "2026-04-16T06:28:47Z"
epic: roadmap-and-portfolio-planning
project: plan
slug: surface-roadmap-progress-in-status-output
spec: roadmap-and-portfolio-planning
status: done
title: Surface roadmap progress in status output
type: story
updated_at: "2026-04-16T06:39:51Z"
---

# Surface roadmap progress in status output

Created: 2026-04-16T06:28:47Z

## Description

Connect roadmap version data to overall project status so users can see what is active and what is parked.
## Acceptance Criteria

- [ ] Status output can group or summarize epics by target version without losing current epic/story progress.

- [ ] Roadmap-aware status stays readable for both small repos and multi-version plans.
## Verification

- go test ./cmd ./internal/planning
## Resources

- [Canonical Spec](../specs/roadmap-and-portfolio-planning.md)
## Notes
