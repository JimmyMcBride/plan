---
created_at: "2026-04-16T06:28:47Z"
epic: plan-quality-and-verification-engine
project: plan
slug: expose-plan-check-command-and-reporting
spec: plan-quality-and-verification-engine
status: done
title: Expose plan check command and reporting
type: story
updated_at: "2026-04-16T06:47:46Z"
---

# Expose plan check command and reporting

Created: 2026-04-16T06:28:47Z

## Description

Add a CLI surface that reports plan quality findings across the workspace in a user-facing way.
## Acceptance Criteria

- [ ] The check command can report findings for project, epic, spec, or story scopes with readable output.

- [ ] Blocking findings produce a clear failing result while non-blocking guidance stays inspectable.
## Verification

- go test ./cmd ./internal/planning
## Resources

- [Canonical Spec](../specs/plan-quality-and-verification-engine.md)
## Notes
