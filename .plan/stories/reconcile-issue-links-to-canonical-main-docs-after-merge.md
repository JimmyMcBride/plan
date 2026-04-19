---
created_at: "2026-04-19T01:53:14Z"
epic: issue-contract-and-planning-link-lifecycle
project: plan
slug: reconcile-issue-links-to-canonical-main-docs-after-merge
spec: issue-contract-and-planning-link-lifecycle
status: done
title: Reconcile issue links to canonical main docs after merge
type: story
updated_at: "2026-04-19T02:55:44Z"
---

# Reconcile issue links to canonical main docs after merge

Created: 2026-04-19T01:53:14Z

## Description

Add reconcile behavior that rewrites `plan`-owned issue sections from pre-merge SHA links to canonical `main` doc links once the planning PR lands.

## Acceptance Criteria

- [ ] Reconcile updates epic/spec links and planning-blocked markers inside `plan`-owned issue sections after merge.

- [ ] Reconcile preserves user edits outside the `plan`-owned issue contract.

## Verification

- Run issue reconcile tests.

## Resources

- [Canonical Spec](../specs/issue-contract-and-planning-link-lifecycle.md)

## Notes
