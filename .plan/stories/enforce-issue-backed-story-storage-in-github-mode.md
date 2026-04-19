---
created_at: "2026-04-19T01:53:14Z"
epic: github-story-backend-and-preflight
project: plan
slug: enforce-issue-backed-story-storage-in-github-mode
spec: github-story-backend-and-preflight
status: done
title: Enforce issue-backed story storage in GitHub mode
type: story
updated_at: "2026-04-19T02:52:42Z"
---

# Enforce issue-backed story storage in GitHub mode

Created: 2026-04-19T01:53:14Z

## Description

Make GitHub mode treat GitHub Issues as story storage so story create, update, list, and show no longer create duplicate local markdown stories when the backend is set to `github`.

## Acceptance Criteria

- [ ] Story operations in GitHub mode avoid writing first-class `.plan/stories/` markdown notes and rely on issue-backed records plus minimal local metadata.

- [ ] Local story mode remains unchanged for repos that do not enable GitHub execution.

## Verification

- Run story backend behavior tests for local and GitHub modes.

## Resources

- [Canonical Spec](../specs/github-story-backend-and-preflight.md)

## Notes
