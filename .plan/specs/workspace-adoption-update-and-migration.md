---
created_at: "2026-04-16T05:33:06Z"
epic: workspace-adoption-update-and-migration
project: plan
slug: workspace-adoption-update-and-migration
status: done
target_version: v2
title: Workspace Adoption, Update, and Migration Spec
type: spec
updated_at: "2026-04-16T06:56:00Z"
---

# Workspace Adoption, Update, and Migration Spec

Created: 2026-04-16T05:33:06Z

## Why

Real projects do not stay pristine forever. `plan` needs a safe way to adopt existing repos and evolve its own workspace without forcing users into manual repair.

## Problem

Fresh init alone is not enough for long-lived usage. The tool needs a clear answer for "how do I start using this in an existing repo?" and "what happens when the workspace format changes later?"

## Goals

- support safe adoption of existing repos
- report pending, missing, or broken workspace state clearly
- keep migrations idempotent and inspectable
- avoid touching non-`plan` files except when explicitly asked

## Non-Goals

- schema-heavy migration machinery
- destructive automatic rewrites of user-authored notes
- remote upgrade coordination

## Constraints

- changes should be limited to `plan`-managed surfaces
- updates must preserve user markdown
- migration state must remain small and readable

## Solution Shape

- expand doctor/update around repairable conditions
- define an `adopt` workflow later for unmanaged repos
- normalize metadata and tool-owned files lazily and safely
- keep migration state in `.plan/.meta/migrations.json`

## Flows

1. User points `plan` at an existing repo.
2. `plan` detects missing or partial workspace state.
3. User runs adoption or update workflow.
4. `plan` creates or repairs only the managed surfaces.
5. `plan doctor` reports current state after repair.

## Data / Interfaces

- migration state file with last run status
- doctor output categories: current, missing, broken, pending
- future adopt command for unmanaged repos

## Risks / Open Questions

- whether adoption should be `plan adopt` or folded into `plan init`
- how much note normalization should happen automatically

## Rollout

- strengthen `doctor` and `update` first
- add adoption flow once the workspace format settles from v1 dogfooding

## Verification

- partial workspace states can be repaired
- user-authored planning files remain intact after update
- doctor output clearly distinguishes missing versus broken state

## Story Breakdown

- [ ] [Detect adoptable workspace states](../stories/detect-adoptable-workspace-states.md)
- [ ] [Add adopt command for existing repos](../stories/add-adopt-command-for-existing-repos.md)
- [ ] [Expand migration tracking and repair coverage](../stories/expand-migration-tracking-and-repair-coverage.md)

## Resources

- [Epic](../epics/workspace-adoption-update-and-migration.md)
- [Product Direction](../PRODUCT.md)

## Notes
