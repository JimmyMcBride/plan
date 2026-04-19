---
created_at: "2026-04-19T01:53:14Z"
epic: dependency-aware-issue-readiness
project: plan
slug: compute-derived-ready-and-blocked-state-from-merge-plus-dependency-closure
spec: dependency-aware-issue-readiness
status: done
title: Compute derived ready and blocked state from merge plus dependency closure
type: story
updated_at: "2026-04-19T03:01:08Z"
---

# Compute derived ready and blocked state from merge plus dependency closure

Created: 2026-04-19T01:53:14Z

## Description

Implement readiness derivation so issue-backed stories stay blocked until the planning PR is merged and all dependency issues are closed, then become ready without manual toggles.

## Acceptance Criteria

- [ ] Issues remain blocked when planning merge state or open dependencies prevent execution.

- [ ] Issues become ready when planning merge state is satisfied and dependency issues are closed.

## Verification

- Run readiness derivation tests.

## Resources

- [Canonical Spec](../specs/dependency-aware-issue-readiness.md)

## Notes
