---
created_at: "2026-04-19T01:16:45Z"
epic: github-story-backend-and-preflight
project: plan
slug: github-story-backend-and-preflight
status: approved
target_version: v7
title: GitHub Story Backend and Preflight Spec
type: spec
updated_at: "2026-04-19T01:54:55Z"
---

# GitHub Story Backend and Preflight Spec

Created: 2026-04-19T01:16:45Z

## Why
If GitHub is going to be the execution backend for stories, `plan` needs a
clear enablement model, hard safety checks, and an unambiguous boundary between
local shaping docs and remote execution units.

## Problem
Right now `plan` assumes stories are local markdown under `.plan/stories/`.
That does not fit the issue-backed model. Without a real backend switch, the
product will either duplicate story truth or create a half-local,
half-GitHub workflow that confuses humans and agents.

## Goals
- add an explicit GitHub story backend
- add `plan github enable` with repo/auth preflight
- require `gh` to be installed and logged in for GitHub mode
- verify the current repo maps to a GitHub repo with Issues enabled
- prevent duplicate local story markdown when GitHub mode is active
- keep only minimal issue metadata locally when stories are GitHub-backed

## Non-Goals
- implementing issue dependency readiness in this spec
- linking planning PRs and permalinks in issue bodies
- installing GitHub Actions workflows
- supporting non-GitHub trackers

## Constraints
- brainstorms, epics, and specs stay canonical in `.plan/`
- GitHub mode must be opt-in
- GitHub mode requires `gh` plus successful `gh auth status`
- story commands must respect the configured backend
- local-first repos that do not enable GitHub mode must keep current behavior

## Solution Shape
- add story backend config with `local` as default and `github` as opt-in
- add `plan github enable` to run preflight and store backend configuration
- preflight checks:
  - `gh` exists on PATH
  - `gh auth status` succeeds
  - current git remote resolves to a GitHub repo
  - GitHub Issues are enabled on the target repo
- in GitHub mode, story create/update/list/show operate on issue-backed records
  instead of local story markdown files
- store only minimal issue metadata locally under `.plan/.meta/`

## Flows
1. User runs `plan github enable`.
2. `plan` checks for `gh`, auth, repo mapping, and Issues support.
3. `plan` stores GitHub backend configuration locally.
4. User creates an execution story from an approved spec.
5. `plan` creates or updates a GitHub issue-backed story and stores minimal
   local metadata instead of writing a local story markdown note.

## Data / Interfaces
- `plan github enable`
- story backend config in `.plan/.meta/`
- minimal issue metadata/index in `.plan/.meta/`
- backend-aware story create/update/list/show commands

## Risks / Open Questions
- whether backend selection should be repo-level only or later support per-epic
  overrides
- how much issue metadata should be cached locally versus fetched on demand
- how to handle migration for repos that already have local stories when GitHub
  mode is enabled later

## Rollout
- ship repo-level GitHub mode first
- keep local story mode as the default
- defer migration ergonomics until the backend model itself is solid
- layer link lifecycle and readiness on top after backend enforcement exists

## Verification
- `plan github enable` fails with clear errors when `gh` is missing or not
  authenticated
- `plan github enable` fails when the repo does not resolve to a GitHub repo or
  Issues are disabled
- GitHub mode prevents local story markdown creation
- local mode continues to create and manage `.plan/stories/` notes unchanged

## Story Breakdown
- [ ] [Add GitHub backend configuration and preflight checks](../stories/add-github-backend-configuration-and-preflight-checks.md)
- [ ] [Enforce issue-backed story storage in GitHub mode](../stories/enforce-issue-backed-story-storage-in-github-mode.md)
- [ ] [Add tests for local and GitHub story backend behavior](../stories/add-tests-for-local-and-github-story-backend-behavior.md)

## Analysis

### Missing Constraints

### Success Criteria Gaps

### Hidden Dependencies

### Risk Gaps

### What/Why vs How Leakage

### Recommended Revisions

## Checklist

## Resources
- [Epic](../epics/github-story-backend-and-preflight.md)
- [Product Direction](../PRODUCT.md)
- [Source Brainstorm](../brainstorms/github-issues-integration.md)

## Notes
