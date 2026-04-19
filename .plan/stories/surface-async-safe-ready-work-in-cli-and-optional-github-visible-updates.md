---
created_at: "2026-04-19T01:53:14Z"
epic: dependency-aware-issue-readiness
project: plan
slug: surface-async-safe-ready-work-in-cli-and-optional-github-visible-updates
spec: dependency-aware-issue-readiness
status: todo
title: Surface async-safe ready work in CLI and optional GitHub-visible updates
type: story
updated_at: "2026-04-19T01:53:14Z"
---

# Surface async-safe ready work in CLI and optional GitHub-visible updates

Created: 2026-04-19T01:53:14Z

## Description

Surface multiple ready issues at once for async work and optionally reflect the same derived readiness into GitHub-visible markers without requiring a GitHub workflow.

## Acceptance Criteria

- [ ] CLI output can show more than one ready issue at a time when blockers do not overlap.

- [ ] Optional GitHub-visible markers use the same derived readiness logic as the CLI instead of a separate automation path.

## Verification

- Run ready-work visibility tests.

## Resources

- [Canonical Spec](../specs/dependency-aware-issue-readiness.md)

## Notes
