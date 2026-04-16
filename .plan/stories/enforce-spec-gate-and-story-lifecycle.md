---
created_at: "2026-04-16T05:46:34Z"
epic: spec-driven-planning-workflow
project: plan
slug: enforce-spec-gate-and-story-lifecycle
spec: spec-driven-planning-workflow
status: todo
title: Enforce spec gate and story lifecycle
type: story
updated_at: "2026-04-16T05:46:34Z"
---

# Enforce spec gate and story lifecycle

Created: 2026-04-16T05:46:34Z

## Description


Strengthen the transition from approved spec to execution-ready stories and status reporting.
## Acceptance Criteria


- [ ] Stories cannot be created from draft specs

- [ ] Story artifacts include acceptance and verification expectations

- [ ] Status reporting reflects epic and story progress cleanly
## Verification


- go test ./internal/planning ./cmd
## Resources


- [Canonical Spec](../specs/spec-driven-planning-workflow.md)
## Notes
