---
created_at: "2026-04-16T06:28:47Z"
epic: dependency-graph-and-ready-work
project: plan
slug: add-dependency-metadata-to-stories
spec: dependency-graph-and-ready-work
status: todo
title: Add dependency metadata to stories
type: story
updated_at: "2026-04-16T06:28:47Z"
---

# Add dependency metadata to stories

Created: 2026-04-16T06:28:47Z

## Description

Let stories declare narrow blocker relationships without leaving the markdown-first model.
## Acceptance Criteria

- [ ] Stories can record blocker story slugs in an inspectable local metadata shape.

- [ ] Dependency metadata stays simple enough for solo developers to maintain by hand when needed.
## Verification

- go test ./internal/planning
## Resources

- [Canonical Spec](../specs/dependency-graph-and-ready-work.md)
## Notes
