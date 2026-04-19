---
created_at: "2026-04-19T01:16:45Z"
project: plan
slug: dependency-aware-issue-readiness
spec: dependency-aware-issue-readiness
title: Dependency-Aware Issue Readiness
target_version: v7
type: epic
updated_at: "2026-04-19T01:16:45Z"
---

# Dependency-Aware Issue Readiness

Created: 2026-04-19T01:16:45Z

## Outcome
Make blocked and ready work obvious for GitHub-backed stories so humans and AI
agents can see execution order and parallel-safe work directly from the issue
contract.

## Why Now
The value of GitHub-backed stories is not just storage. It is execution clarity.
Without visible blockers and derived readiness, issue-backed stories still
leave users guessing what comes next or what can run in parallel.

## Shape

### Appetite
One `v7` epic: dependency encoding, derived ready state, async-safe parallel
lanes, and optional reconcile-driven updates. No required GitHub workflow.

### Outcome
Issue-backed stories show blockers clearly, `plan` can compute ready work from
planning-merge state plus dependency closure, and multiple ready issues can be
worked in parallel when they do not conflict.

### Scope Boundary
- dependency representation for issue-backed stories
- derived blocked and ready state
- multiple ready issues when blockers do not overlap
- CLI visibility into next-ready work
- optional GitHub-visible readiness updates driven by the same CLI logic

### Out of Scope
- mandatory GitHub Actions setup
- Jira/Linear-style scheduling features
- project-board orchestration
- invisible dependency logic that exists only in `.plan`

### Success Signal
Humans and agents can open the issue set and reliably tell what is blocked,
what is ready now, and what can proceed in parallel without guessing.

## Scope Boundary
- dependency contract for issue-backed stories
- readiness derivation rules
- CLI reconcile/status surfaces
- optional GitHub-visible updates such as labels or body markers

## Spec
- [Draft Spec](../specs/dependency-aware-issue-readiness.md)

## Resources
- [Product Direction](../PRODUCT.md)
- [Roadmap](../ROADMAP.md)
- [Source Brainstorm](../brainstorms/github-issues-integration.md)

## Progress
- Target version: `v7`
- Status: planned

## Notes
Readiness should be derived, not manually toggled. Automation may call the same
logic later, but the logic itself belongs inside `plan`.
