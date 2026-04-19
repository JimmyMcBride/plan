---
created_at: "2026-04-19T01:53:14Z"
epic: github-story-backend-and-preflight
project: plan
slug: add-tests-for-local-and-github-story-backend-behavior
spec: github-story-backend-and-preflight
status: done
title: Add tests for local and GitHub story backend behavior
type: story
updated_at: "2026-04-19T02:52:42Z"
---

# Add tests for local and GitHub story backend behavior

Created: 2026-04-19T01:53:14Z

## Description

Expand tests so backend selection, preflight, and story storage behavior stay correct across local-only repos and GitHub-enabled repos.

## Acceptance Criteria

- [ ] Tests cover successful GitHub enablement plus failure paths for missing `gh`, failed auth, and unsupported repo state.

- [ ] Tests catch regressions where GitHub mode accidentally creates duplicate local story notes or breaks local story mode.

## Verification

- Run local and GitHub story backend test suites.

## Resources

- [Canonical Spec](../specs/github-story-backend-and-preflight.md)

## Notes
