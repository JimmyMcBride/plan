---
created_at: "2026-04-16T05:33:06Z"
project: plan
slug: brain-interop-and-planning-imports
spec: brain-interop-and-planning-imports
target_version: v3
title: Brain Interop and Planning Imports
type: epic
updated_at: "2026-04-16T05:33:06Z"
---

# Brain Interop and Planning Imports

Created: 2026-04-16T05:33:06Z

## Outcome

Let `plan` adopt or import planning material from `brain` cleanly while preserving the strict product boundary between planning and memory.

## Why Now

`brain` and `plan` should work together, but only after `plan` has proven its own native model. `v3` is the first sensible time to bridge them.

## Scope Boundary

- import existing `brain` planning notes
- preserve links between imported artifacts and source notes
- clarify the product boundary between `brain` and `plan`

Not in scope:

- merging the two tools
- moving memory or search features into `plan`
- cloud sync between tools

## Spec

- [Draft Spec](../specs/brain-interop-and-planning-imports.md)

## Resources

- [Product Direction](../PRODUCT.md)
- [Reference Analysis](../research/reference-analysis.md)

## Progress

- Target version: `v3`
- Status: planned

## Notes

Interop should strengthen both tools without muddying their responsibilities.
