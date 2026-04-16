---
created_at: "2026-04-16T05:46:33Z"
epic: spec-driven-planning-workflow
project: plan
slug: improve-epic-promotion-and-spec-seeding
spec: spec-driven-planning-workflow
status: todo
title: Improve epic promotion and spec seeding
type: story
updated_at: "2026-04-16T05:46:33Z"
---

# Improve epic promotion and spec seeding

Created: 2026-04-16T05:46:33Z

## Description


Make brainstorm promotion reliably create an epic and a useful seeded draft spec.
## Acceptance Criteria


- [ ] Promotion creates a linked epic and canonical spec

- [ ] Seeded spec captures useful problem and goal material from the brainstorm

- [ ] Source provenance back to the brainstorm remains visible
## Verification


- go test ./internal/planning
## Resources


- [Canonical Spec](../specs/spec-driven-planning-workflow.md)
## Notes
