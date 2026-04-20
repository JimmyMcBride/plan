---
created_at: "2026-04-20T00:07:47Z"
project: plan
slug: reopen-review-and-roadmap-parking
spec: reopen-review-and-roadmap-parking
title: Reopen Review and Roadmap Parking
type: epic
updated_at: "2026-04-20T00:07:47Z"
---

# Reopen Review and Roadmap Parking

Created: 2026-04-20T00:07:47Z

## Outcome

Let users safely move backward, review stale downstream stages, and park good
ideas that are valid but too early.

## Why Now

Guided systems fail if they are either too rigid or too chaotic. `plan` needs a
safe way to reopen earlier thinking and a durable place to put good ideas that
should not enter the active feature yet.

## Shape

### Appetite

Medium. Add safety rails, not a second workflow system.

### Outcome

Users can reopen earlier stages without silent damage, downstream stages become
`needs review`, and good-but-early ideas land in roadmap parking instead of
being forgotten.

### Scope Boundary

- explicit reopen-stage flow
- downstream `needs review` markers
- quick review checkpoints before moving forward again
- roadmap parking writes with durable metadata
- visible impact summary before upstream edits

### Out of Scope

- auto-deleting downstream work
- hard-blocking all stale stages forever
- a separate backlog artifact outside `ROADMAP.md`

### Success Signal

Backward jumps feel safe and obvious, and parked ideas remain useful later.

## Scope Boundary

- reopen guidance and impact summary
- stale downstream markers and review flow
- roadmap parking entry structure

Not in scope:

- broad portfolio management
- dependency graph product work
- hidden automatic rewrites of downstream artifacts

## Spec

- [Draft Spec](../specs/reopen-review-and-roadmap-parking.md)

## Resources

- [Brainstorm](../brainstorms/guided-planning-system.md)
- [Product Direction](../PRODUCT.md)
- [Roadmap](../ROADMAP.md)

## Progress

- Target version: `v8`
- Status: planned

## Notes

This epic should make guided planning flexible without collapsing trust in the
artifact chain.
