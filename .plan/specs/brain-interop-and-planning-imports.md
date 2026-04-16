---
created_at: "2026-04-16T05:33:06Z"
epic: brain-interop-and-planning-imports
project: plan
slug: brain-interop-and-planning-imports
status: done
target_version: v3
title: Brain Interop and Planning Imports Spec
type: spec
updated_at: "2026-04-16T07:09:59Z"
---

# Brain Interop and Planning Imports Spec

Created: 2026-04-16T05:33:06Z

## Why

Many early users of `plan` will already have planning material in `brain`. `plan` should provide a clean migration path without trying to absorb everything `brain` does.

## Problem

Without an interop story, users may need to re-create valuable planning work by hand or keep duplicate planning systems around too long.

## Goals

- import epic/spec/story-relevant material from `brain`
- preserve source provenance
- keep `brain` focused on memory/context and `plan` focused on planning
- avoid duplicate ownership of planning notes after import

## Non-Goals

- importing all `brain` context
- pulling in search/session systems
- creating a unified super-tool

## Constraints

- imports must be explicit
- imported material should be inspectable before final promotion
- provenance should remain visible in the resulting notes

## Solution Shape

- add import flows for `brain` planning notes
- map `brain` structures into the `plan` canonical model
- link imported artifacts back to their source where helpful
- keep interop optional and local

## Flows

1. User points `plan` at a `brain` workspace or notes.
2. `plan` previews importable planning material.
3. User selects what to import.
4. `plan` writes new `plan` notes with provenance links.

## Data / Interfaces

- source metadata on imported notes
- preview output for import candidates
- explicit import commands or subcommands

## Risks / Open Questions

- how much automatic mapping is safe
- whether import should create fresh notes or attempt in-place conversion

## Rollout

- begin with one-way import from `brain` to `plan`
- defer richer synchronization until much later, if ever

## Verification

- import preserves meaning and provenance
- imported notes align with the `Epic -> Spec -> Story` model
- `plan` still does not depend on `brain` for core behavior

## Story Breakdown

- [ ] [Inspect brain workspaces for import candidates](../stories/inspect-brain-workspaces-for-import-candidates.md)
- [ ] [Import brain planning notes into plan artifacts](../stories/import-brain-planning-notes-into-plan-artifacts.md)
- [ ] [Preserve import provenance and review flow](../stories/preserve-import-provenance-and-review-flow.md)

## Resources

- [Epic](../epics/brain-interop-and-planning-imports.md)
- [Reference Analysis](../research/reference-analysis.md)

## Notes
