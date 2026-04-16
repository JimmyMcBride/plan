---
created_at: "2026-04-16T05:33:06Z"
epic: core-workspace-and-artifact-system
project: plan
slug: core-workspace-and-artifact-system
status: approved
target_version: v1
title: Core Workspace and Artifact System Spec
type: spec
updated_at: "2026-04-16T05:46:33Z"
---

# Core Workspace and Artifact System Spec

Created: 2026-04-16T05:33:06Z

## Why

`plan` needs a durable and predictable home inside every repo. The workspace model has to be simple enough for new users and stable enough to survive future upgrades.

## Problem

Right now the repo has early planning notes, but not a fully defined and durable `plan` workspace contract. Without that contract, later workflows will drift.

## Goals

- define the canonical `.plan/` layout
- keep user-authored planning files separate from tool-owned metadata
- support `init`, `doctor`, and `update` as the base lifecycle
- make workspace upgrades inspectable and idempotent

## Non-Goals

- complex database migrations
- cloud sync
- dependency graph modeling
- context or memory management

## Constraints

- local-first only
- markdown-first for planning artifacts
- `.meta/` reserved for tool-owned state
- migration behavior must stay small and understandable

## Solution Shape

- use `.plan/PROJECT.md` and `.plan/ROADMAP.md` as top-level planning anchors
- store epics, specs, and stories in dedicated directories
- store workspace and migration state in `.plan/.meta/*.json`
- use `plan doctor` to report current, missing, or broken state

## Flows

1. User runs `plan init`.
2. `plan` creates missing workspace directories and root planning files.
3. User works in the workspace.
4. `plan doctor` validates health.
5. `plan update` repairs or normalizes tool-managed surfaces.

## Data / Interfaces

- `workspace.json`
  - schema version
  - planning model
  - created/updated timestamps
- `migrations.json`
  - schema version
  - known migrations
  - last run status

## Risks / Open Questions

- how much metadata belongs in frontmatter versus `.meta/`
- whether future import/adopt flows need additional workspace markers

## Rollout

- ship in `v1`
- use `plan` itself as the first dogfood workspace
- keep migration state minimal until real upgrade cases appear

## Verification

- `plan init --project .` creates the expected layout
- `plan doctor --project .` reports `current` after init
- `plan update --project .` repairs missing tool-managed files without damaging user notes

## Story Breakdown

- [ ] [Define .plan workspace contract](../stories/define-plan-workspace-contract.md)
- [ ] [Harden workspace metadata and repair lifecycle](../stories/harden-workspace-metadata-and-repair-lifecycle.md)
- [ ] [Expand workspace test coverage](../stories/expand-workspace-test-coverage.md)

## Resources

- [Epic](../epics/core-workspace-and-artifact-system.md)
- [Product Direction](../PRODUCT.md)

## Notes
