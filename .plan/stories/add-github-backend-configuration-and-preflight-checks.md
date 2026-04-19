---
created_at: "2026-04-19T01:53:14Z"
epic: github-story-backend-and-preflight
project: plan
slug: add-github-backend-configuration-and-preflight-checks
spec: github-story-backend-and-preflight
status: done
title: Add GitHub backend configuration and preflight checks
type: story
updated_at: "2026-04-19T02:46:37Z"
---

# Add GitHub backend configuration and preflight checks

Created: 2026-04-19T01:53:14Z

## Description

Add the repo-level GitHub story backend setting plus the preflight path that validates `gh`, auth, repo mapping, and GitHub Issues support before GitHub mode can be enabled.

## Acceptance Criteria

- [ ] `plan github enable` stores backend configuration only after `gh`, auth, repo mapping, and Issues checks pass.

- [ ] Failure cases return clear guidance when `gh` is missing, not logged in, or the target repo cannot support issue-backed stories.

## Verification

- Run GitHub backend preflight tests.

## Resources

- [Canonical Spec](../specs/github-story-backend-and-preflight.md)

## Notes
