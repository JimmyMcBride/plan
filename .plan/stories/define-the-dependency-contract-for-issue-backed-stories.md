---
created_at: "2026-04-19T01:53:14Z"
epic: dependency-aware-issue-readiness
project: plan
slug: define-the-dependency-contract-for-issue-backed-stories
spec: dependency-aware-issue-readiness
status: todo
title: Define the dependency contract for issue-backed stories
type: story
updated_at: "2026-04-19T01:53:14Z"
---

# Define the dependency contract for issue-backed stories

Created: 2026-04-19T01:53:14Z

## Description

Define how blockers and dependencies are represented in issue-backed stories so they stay visible from the issue itself and remain stable enough for `plan` to compute readiness.

## Acceptance Criteria

- [ ] The dependency contract is visible in the issue body or another GitHub-visible surface, not hidden only in local metadata.

- [ ] Dependency references are structured enough for `plan` to parse and use for readiness derivation.

## Verification

- Run dependency contract parsing tests.

## Resources

- [Canonical Spec](../specs/dependency-aware-issue-readiness.md)

## Notes
