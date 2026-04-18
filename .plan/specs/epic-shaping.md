---
created_at: "2026-04-17T20:19:39Z"
epic: epic-shaping
project: plan
slug: epic-shaping
status: done
target_version: v5
title: Epic Shaping Spec
type: spec
updated_at: "2026-04-17T20:34:02Z"
---

# Epic Shaping Spec

Created: 2026-04-17T20:19:39Z

## Why

Epics should be more than named containers. They should capture the shaped bet
that makes the later spec stronger.

## Problem

The current epic model lacks appetite, explicit out-of-scope decisions, and a
clear success signal, which leaves a gap between brainstorm work and spec work.

## Goals

- add a durable `## Shape` section to epic notes
- ship `plan epic shape`
- make shaped epic output stable and easy to inspect
- keep specs as the canonical execution contract

## Non-Goals

- replacing specs with epics as the main source of truth
- adding new hierarchy layers
- tying shaping to external trackers

## Constraints

- epic shaping must stay additive to the current note structure
- the command should update notes safely and idempotently
- shape output should be understandable to both humans and agents
- the simple `epic create` path should remain valid without immediate shaping

## Solution Shape

- extend the epic template with a fixed `## Shape` section
- add a structured `epic shape` command for appetite, outcome, scope boundary,
  out of scope, and success signal
- use note update helpers that preserve the rest of the epic content

## Flows

1. User creates or promotes an epic.
2. User runs `plan epic shape <epic-slug>`.
3. `plan` records appetite, outcome, scope boundary, out-of-scope decisions, and
   a success signal.
4. The shaped epic informs the later spec and roadmap conversations.

## Data / Interfaces

- epic template additions
- an `EpicShapeInput` in planning code
- the `plan epic shape` command

## Risks / Open Questions

- whether shaped fields should also update the older top-level epic sections
- how much shape data should later seed specs automatically

## Rollout

- land the note schema and command in `v5`
- verify note writes and output stability in tests
- consider deeper spec seeding only after the basic flow proves useful

## Verification

- new epics include the `## Shape` headings
- `plan epic shape` updates the correct sections without damaging the rest of the
  note
- tests cover first-run and rerun behavior

## Story Breakdown

- [x] [Add epic shape note schema](../stories/add-epic-shape-note-schema.md)
- [x] [Implement epic shape command flow](../stories/implement-epic-shape-command-flow.md)
- [x] [Add epic shape tests and check coverage](../stories/add-epic-shape-tests-and-check-coverage.md)

## Analysis

### Missing Constraints

### Success Criteria Gaps

### Hidden Dependencies

### Risk Gaps

### What/Why vs How Leakage

### Recommended Revisions

## Resources

- [Epic](../epics/epic-shaping.md)
- [Product Direction](../PRODUCT.md)

## Notes
