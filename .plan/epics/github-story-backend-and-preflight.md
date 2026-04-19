---
created_at: "2026-04-19T01:16:45Z"
project: plan
slug: github-story-backend-and-preflight
spec: github-story-backend-and-preflight
title: GitHub Story Backend and Preflight
target_version: v7
type: epic
updated_at: "2026-04-19T01:16:45Z"
---

# GitHub Story Backend and Preflight

Created: 2026-04-19T01:16:45Z

## Outcome
Add an opt-in GitHub execution backend where stories live in GitHub Issues,
while brainstorms, epics, and specs remain canonical local markdown under
`.plan/`.

## Why Now
The product direction is no longer "mirror local stories into GitHub." If issue
execution is real, `plan` needs a clean backend boundary, explicit enablement,
and hard preflight checks before it starts treating GitHub as story storage.

## Shape

### Appetite
One contained `v7` epic: backend selection, preflight, and issue-backed story
storage semantics. No dependency orchestration or workflow automation here.

### Outcome
Users can enable GitHub mode safely, verify repo/auth prerequisites up front,
and create execution stories without duplicating them as local markdown notes.

### Scope Boundary
- `plan github enable`
- `gh` presence and auth preflight
- GitHub repo and Issues capability checks
- explicit story backend selection: `local` or `github`
- issue-backed story storage with minimal local metadata/index

### Out of Scope
- issue dependency readiness
- planning PR link lifecycle
- GitHub workflow installation
- Jira/Linear or non-GitHub trackers

### Success Signal
GitHub mode can be enabled intentionally, fails safely when prerequisites are
missing, and prevents duplicate story truth across `.plan` and GitHub Issues.

## Scope Boundary
- repo-level GitHub mode enablement
- CLI preflight checks and error messages
- backend-aware story create/update/list/show behavior
- local metadata for issue-backed stories

## Spec
- [Draft Spec](../specs/github-story-backend-and-preflight.md)

## Resources
- [Product Direction](../PRODUCT.md)
- [Roadmap](../ROADMAP.md)
- [Source Brainstorm](../brainstorms/github-issues-integration.md)

## Progress
- Target version: `v7`
- Status: planned

## Notes
Strict default: if GitHub mode is enabled, stories should not also exist as
first-class local markdown notes.
