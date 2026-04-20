---
created_at: "2026-04-20T00:07:47Z"
epic: reopen-review-and-roadmap-parking
project: plan
slug: reopen-review-and-roadmap-parking
status: approved
target_version: v8
title: Reopen Review and Roadmap Parking Spec
type: spec
updated_at: "2026-04-20T03:04:50Z"
---

# Reopen Review and Roadmap Parking Spec

Created: 2026-04-20T00:07:47Z

## Why

Guided planning needs flexibility, but not at the cost of hidden inconsistency.

## Problem

Users will sometimes realize they need to reopen an earlier stage. Without an
explicit reopen and review model, downstream stages can quietly drift out of
date. At the same time, good ideas that emerge during planning need a durable
home when they are real but premature.

## Goals

- allow explicit backward jumps to earlier stages
- show impact before upstream edits
- mark downstream stages as `needs review` when confidence drops
- require a lightweight review checkpoint before moving forward again
- park good-but-early ideas in `ROADMAP.md` with useful metadata

## Non-Goals

- silently mutating downstream artifacts
- hard-blocking the user behind heavy review ceremony
- introducing a separate backlog system

## Constraints

- downstream review should feel lightweight
- stale state must be visible, not hidden
- roadmap parking remains inside `ROADMAP.md`
- users stay in control of whether to reopen and how much to revisit

## Solution Shape

- add an explicit reopen-stage action
- before reopening, show:
  - stage to reopen
  - downstream stages likely affected
  - one recommended path plus 1-2 alternatives
- mark later stages as `needs review`
- before continuing forward, run a short review checkpoint on stale stages
- when a good-but-early idea appears, suggest parking it and on confirmation
  write:
  - title
  - value or outcome
  - why parked now
  - unlock condition
  - source reference

## Flows

1. User decides the current stage flow needs an earlier change.
2. `plan` offers `reopen stage`.
3. `plan` shows impact summary and recommended path.
4. User confirms reopen.
5. Upstream stage is reopened; downstream stages become `needs review`.
6. Later, when the user moves forward again, `plan` runs quick review
   checkpoints in order.
7. During any stage, if an idea is good but too early, `plan` recommends
   parking it in `ROADMAP.md`.

## Data / Interfaces

- stage freshness metadata in session state
- `needs review` marker for downstream stages
- roadmap parking entry structure
- reopen-stage impact summary output

## Risks / Open Questions

- how aggressive `plan` should be when deciding something is parking-lot worthy
- how much downstream content should be shown in review checkpoints

## Rollout

- land reopen impact summaries and `needs review` markers first
- add review checkpoints next
- add roadmap parking writes and source references last

## Verification

- reopening an earlier stage shows impact before mutation
- downstream stages become `needs review`
- moving forward again triggers lightweight downstream review
- parked ideas are written to `ROADMAP.md` with value, reason, unlock, and source

## Story Breakdown

- [ ] [Add Reopen Impact Summary And Needs Review Markers](https://github.com/JimmyMcBride/plan/issues/19)
- [ ] [Implement Downstream Review Checkpoints](https://github.com/JimmyMcBride/plan/issues/20)
- [ ] [Add Roadmap Parking Writes And Source Links](https://github.com/JimmyMcBride/plan/issues/21)

## Analysis

### Missing Constraints

### Success Criteria Gaps

### Hidden Dependencies

### Risk Gaps

### What/Why vs How Leakage

### Recommended Revisions

## Checklist

## Resources

- [Epic](../epics/reopen-review-and-roadmap-parking.md)
- [Brainstorm](../brainstorms/guided-planning-system.md)
- [Product Direction](../PRODUCT.md)

## Notes

Recommendation locked during brainstorming: downstream stale stages should
require review, not a hard block, and roadmap parking should live inside
`ROADMAP.md`.
