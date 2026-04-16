---
created_at: "2026-04-16T06:28:47Z"
epic: roadmap-and-portfolio-planning
project: plan
slug: define-roadmap-version-structure-and-parsing
spec: roadmap-and-portfolio-planning
status: done
title: Define roadmap version structure and parsing
type: story
updated_at: "2026-04-16T06:33:37Z"
---

# Define roadmap version structure and parsing

Created: 2026-04-16T06:28:47Z

## Description

Make ROADMAP.md version sections and summaries explicit so plan can read them reliably.
## Acceptance Criteria

- [ ] Roadmap parsing reads version goals, summaries, epic lists, and parking lot notes from markdown.

- [ ] Parsing keeps version order stable and tolerates empty sections without data loss.
## Verification

- go test ./internal/planning
## Resources

- [Canonical Spec](../specs/roadmap-and-portfolio-planning.md)
## Notes
