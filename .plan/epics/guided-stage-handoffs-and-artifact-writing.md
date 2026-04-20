---
created_at: "2026-04-20T00:07:47Z"
project: plan
slug: guided-stage-handoffs-and-artifact-writing
spec: guided-stage-handoffs-and-artifact-writing
title: Guided Stage Handoffs and Artifact Writing
type: epic
updated_at: "2026-04-20T00:07:47Z"
---

# Guided Stage Handoffs and Artifact Writing

Created: 2026-04-20T00:07:47Z

## Outcome

Walk users stage-by-stage from brainstorm through epic, spec, and story
creation with explicit checkpoints and durable artifact updates.

## Why Now

Even if brainstorm guidance improves, `plan` will still feel fragmented unless
the user can move forward through later stages without dropping into a loose
collection of disconnected commands.

## Shape

### Appetite

Large enough to cover the main stage chain, but still constrained to one clear
guided path.

### Outcome

Users can continue from brainstorm into epic, spec, and stories through the
same guided flow, with each stage ending in a recap and a clear next action.

### Scope Boundary

- guided progression from brainstorm to epic to spec to stories
- artifact writes at stage checkpoints
- recap and decision menu at every stage
- improved current commands instead of a parallel guided command family
- story creation stage that still produces execution-ready outputs

### Out of Scope

- stale-review and reopen logic
- repo-wide portfolio planning
- hosted collaboration

### Success Signal

The user can stay in one coherent guided flow from rough vision to first story
set without feeling thrown back into raw artifact management.

## Scope Boundary

- improve current stage commands under the hood
- recap structure shared across stages
- stage transition prompts and artifact write points
- stop-for-now next-action summaries

Not in scope:

- upstream reopen handling
- roadmap parking writes
- brand new top-level guided command family

## Spec

- [Draft Spec](../specs/guided-stage-handoffs-and-artifact-writing.md)

## Resources

- [Brainstorm](../brainstorms/guided-planning-system.md)
- [Product Direction](../PRODUCT.md)
- [Roadmap](../ROADMAP.md)

## Progress

- Target version: `v8`
- Status: planned

## Notes

This epic is where guided planning stops being a better brainstorm and becomes
an end-to-end planning system.
