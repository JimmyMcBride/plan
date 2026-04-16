---
created_at: "2026-04-16T06:28:47Z"
epic: roadmap-and-portfolio-planning
project: plan
slug: add-roadmap-cli-helpers-for-version-views
spec: roadmap-and-portfolio-planning
status: done
title: Add roadmap CLI helpers for version views
type: story
updated_at: "2026-04-16T06:34:58Z"
---

# Add roadmap CLI helpers for version views

Created: 2026-04-16T06:28:47Z

## Description

Expose roadmap views that make version planning readable without hand-editing every query.
## Acceptance Criteria

- [ ] Roadmap commands can show version-scoped summaries and ordered epics from ROADMAP.md.

- [ ] Version-focused views preserve the lightweight markdown-first roadmap model.
## Verification

- go test ./cmd ./internal/planning
## Resources

- [Canonical Spec](../specs/roadmap-and-portfolio-planning.md)
## Notes
