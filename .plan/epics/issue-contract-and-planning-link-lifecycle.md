---
created_at: "2026-04-19T01:16:45Z"
project: plan
slug: issue-contract-and-planning-link-lifecycle
spec: issue-contract-and-planning-link-lifecycle
title: Issue Contract and Planning Link Lifecycle
target_version: v7
type: epic
updated_at: "2026-04-19T01:16:45Z"
---

# Issue Contract and Planning Link Lifecycle

Created: 2026-04-19T01:16:45Z

## Outcome
Define the issue body contract and lifecycle that links GitHub-backed stories
to real local epic/spec docs before merge and after merge.

## Why Now
Issue-backed stories are only useful if the issue itself points to trustworthy
planning docs. Planning often happens on a branch before the docs land on
`main`, so the contract must handle branch-state links without making them
fragile or misleading.

## Shape

### Appetite
One `v7` epic: issue body schema, planning PR links, SHA permalinks before
merge, and reconcile to canonical `main` links after merge.

### Outcome
Every GitHub-backed story issue clearly shows where the shaping docs live, what
planning PR it depends on, and how those links normalize after merge.

### Scope Boundary
- visible issue body sections for epic/spec context
- machine-readable metadata block in the issue body
- planning PR links
- commit-SHA permalink strategy before merge
- reconcile flow that rewrites links to canonical `main` docs after merge

### Out of Scope
- issue dependency readiness rules
- board or milestone automation
- exporting brainstorms as issues
- webhook-driven realtime sync

### Success Signal
Users can create issue-backed stories from a planning branch and still trust the
links in the issue body before and after the planning PR merges.

## Scope Boundary
- issue body rendering
- planning branch doc links
- post-merge reconcile behavior
- preserving readable issue text while keeping machine metadata stable

## Spec
- [Draft Spec](../specs/issue-contract-and-planning-link-lifecycle.md)

## Resources
- [Product Direction](../PRODUCT.md)
- [Roadmap](../ROADMAP.md)
- [Source Brainstorm](../brainstorms/github-issues-integration.md)

## Progress
- Target version: `v7`
- Status: planned

## Notes
Branch-name links should not be treated as canonical. Commit-SHA permalinks
before merge, then canonical `main` links after merge.
