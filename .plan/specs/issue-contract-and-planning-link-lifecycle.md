---
created_at: "2026-04-19T01:16:45Z"
epic: issue-contract-and-planning-link-lifecycle
project: plan
slug: issue-contract-and-planning-link-lifecycle
status: done
target_version: v7
title: Issue Contract and Planning Link Lifecycle Spec
type: spec
updated_at: "2026-04-19T02:55:44Z"
---

# Issue Contract and Planning Link Lifecycle Spec

Created: 2026-04-19T01:16:45Z

## Why
If stories live in GitHub Issues, the issue body becomes part of the execution
contract. It must point to real planning docs even when those docs still live
on an unmerged planning branch.

## Problem
Branch planning happens before merge, but issue-backed execution cannot rely on
broken or ambiguous links. Without a clear link lifecycle, users will either
avoid creating issues until late, or they will create issues that point at docs
which drift, disappear, or cannot be trusted after merge.

## Goals
- define a human-readable and machine-readable issue contract
- link each issue-backed story to canonical epic/spec markdown files
- include the planning PR link in the issue when work is created before merge
- use pushed commit-SHA permalinks before merge
- reconcile links to canonical `main` doc links after merge
- make planning-merge state visible from the issue itself

## Non-Goals
- exporting brainstorms, epics, or specs as GitHub issues
- relying on branch-name links as canonical references
- full remote-to-local field sync from arbitrary issue edits
- milestone and project-board automation

## Constraints
- issue links must resolve to real files before merge and after merge
- commit-SHA links should be the default pre-merge doc reference
- issue body contract must remain readable to humans, not hidden in local state
- remote issue edits must not silently rewrite canonical local planning docs
- reconcile logic should preserve issue edits outside the `plan`-owned sections

## Solution Shape
- add an issue body template with:
  - summary/description
  - acceptance criteria
  - verification
  - epic link
  - spec link
  - planning PR link when relevant
  - blocker / dependency section
  - optional async notes
- include a machine-readable `plan` metadata block in the issue body
- before merge, publish epic/spec links as GitHub commit-SHA permalinks
- after merge, run reconcile to rewrite doc links to canonical `main` links and
  clear planning-blocked markers when appropriate

## Flows
1. User shapes work on a planning branch and pushes it.
2. User opens or references a planning PR.
3. `plan` creates or updates an issue-backed story.
4. The issue body contains epic/spec SHA permalinks plus the planning PR link.
5. Planning PR merges.
6. `plan github reconcile` rewrites epic/spec links to canonical `main` links
   and updates planning-blocked markers.

## Data / Interfaces
- issue body schema for issue-backed stories
- machine-readable metadata block embedded in the issue body
- planning PR reference
- doc reference mode: commit SHA before merge, `main` after reconcile
- `plan github reconcile`

## Risks / Open Questions
- whether reconcile should require an explicit PR number or infer it from branch
  and repo context
- how aggressively `plan` should rewrite existing issue bodies after merge
- how much issue body structure should be configurable without breaking agent
  readability

## Rollout
- ship the issue contract alongside the GitHub story backend
- support planning-branch publish before merge
- add reconcile to normalize links after merge
- keep workflow automation optional and secondary to the CLI path

## Verification
- issue-backed stories created from a planning branch include real SHA permalink
  doc links and planning PR references
- after reconcile, issue links point to canonical `main` docs
- issue bodies preserve non-`plan` user edits outside `plan`-owned sections
- no branch-name link is required as the canonical reference path

## Story Breakdown
- [ ] [Define the issue body contract and metadata block](../stories/define-the-issue-body-contract-and-metadata-block.md)
- [ ] [Publish planning-branch doc links and planning PR references](../stories/publish-planning-branch-doc-links-and-planning-pr-references.md)
- [ ] [Reconcile issue links to canonical `main` docs after merge](../stories/reconcile-issue-links-to-canonical-main-docs-after-merge.md)

## Analysis

### Missing Constraints

### Success Criteria Gaps

### Hidden Dependencies

### Risk Gaps

### What/Why vs How Leakage

### Recommended Revisions

## Checklist

## Resources
- [Epic](../epics/issue-contract-and-planning-link-lifecycle.md)
- [Product Direction](../PRODUCT.md)
- [Source Brainstorm](../brainstorms/github-issues-integration.md)

## Notes
