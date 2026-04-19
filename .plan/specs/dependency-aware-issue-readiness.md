---
created_at: "2026-04-19T01:16:45Z"
epic: dependency-aware-issue-readiness
project: plan
slug: dependency-aware-issue-readiness
status: done
target_version: v7
title: Dependency-Aware Issue Readiness Spec
type: spec
updated_at: "2026-04-19T03:01:09Z"
---

# Dependency-Aware Issue Readiness Spec

Created: 2026-04-19T01:16:45Z

## Why
If GitHub Issues are the execution backend for stories, they need to do more
than exist. They need to expose sequencing and parallel-ready work clearly
enough for humans and agents to act without hidden local knowledge.

## Problem
Today issue trackers easily devolve into flat queues. That breaks the promise
of GitHub-backed stories for `plan`. Users need to know what is blocked, what
becomes ready after merge, and what can safely proceed in parallel. If
dependencies live only in local metadata, the issue itself stays ambiguous.

## Goals
- encode story dependencies in a way visible from the issue itself
- derive issue readiness from planning-merge state plus dependency closure
- allow multiple ready issues at once when blockers do not overlap
- surface ready and blocked work in `plan` without requiring a GitHub workflow
- optionally reflect derived readiness back into GitHub using the same CLI logic

## Non-Goals
- mandatory GitHub Actions or webhook automation
- full project scheduling or enterprise resource planning
- hidden dependency logic that only `plan` can see
- milestone planning as a prerequisite for dependency-aware readiness

## Constraints
- readiness must be derived, not manually toggled
- planning PR merge can block readiness before dependencies are even considered
- dependency information must be visible in the issue body or another
  GitHub-visible surface
- the same readiness logic must work from the CLI without any workflow setup
- optional automation must call the same underlying `plan` logic, not invent a
  second implementation

## Solution Shape
- define a dependency contract for issue-backed stories
- default to a `plan`-owned dependency section in the issue body, with optional
  native GitHub dependency mirroring later if available and reliable
- derive issue state from:
  - planning merge status
  - dependency issue closure
- surface readiness in `plan status` and related GitHub-aware views
- optionally update GitHub-visible markers such as labels, body sections, or
  comments from explicit reconcile runs

## Flows
1. User creates issue-backed stories with blockers/dependencies.
2. `plan` records dependency metadata in the issue contract.
3. `plan github reconcile` reads planning merge state plus dependency closure.
4. `plan` computes which issues are blocked, which are ready, and which are
   parallel-safe.
5. Users or agents pick any ready issue, while blocked issues clearly point to
   what must close first.

## Data / Interfaces
- issue dependency section inside the issue contract
- optional minimal local readiness cache/index
- `plan github reconcile`
- GitHub-aware status output and ready-work reporting

## Risks / Open Questions
- how much visible issue-body dependency structure is acceptable before it feels
  noisy
- whether labels are enough for readiness visibility or whether issue body
  markers should stay primary
- how to represent parallel-safe work without implying stronger scheduling than
  the system can actually guarantee

## Rollout
- ship dependency encoding and CLI-derived readiness first
- keep GitHub workflow automation optional
- add GitHub-visible readiness updates only through the same reconcile path
- revisit native GitHub dependency mirroring only after the body contract proves
  stable

## Verification
- blocked issues clearly indicate which dependencies remain open
- ready issues become visible when planning merge and dependency closure allow
  execution
- multiple ready issues can be surfaced at once for async work
- no GitHub Actions setup is required to compute readiness correctly

## Story Breakdown
- [ ] [Define the dependency contract for issue-backed stories](../stories/define-the-dependency-contract-for-issue-backed-stories.md)
- [ ] [Compute derived ready and blocked state from merge plus dependency closure](../stories/compute-derived-ready-and-blocked-state-from-merge-plus-dependency-closure.md)
- [ ] [Surface async-safe ready work in CLI and optional GitHub-visible updates](../stories/surface-async-safe-ready-work-in-cli-and-optional-github-visible-updates.md)

## Analysis

### Missing Constraints

### Success Criteria Gaps

### Hidden Dependencies

### Risk Gaps

### What/Why vs How Leakage

### Recommended Revisions

## Checklist

## Resources
- [Epic](../epics/dependency-aware-issue-readiness.md)
- [Product Direction](../PRODUCT.md)
- [Source Brainstorm](../brainstorms/github-issues-integration.md)

## Notes
