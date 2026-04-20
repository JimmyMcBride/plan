---
created_at: "2026-04-20T00:07:47Z"
project: plan
slug: guided-session-engine-and-resume
spec: guided-session-engine-and-resume
title: Guided Session Engine and Resume
type: epic
updated_at: "2026-04-20T00:07:47Z"
---

# Guided Session Engine and Resume

Created: 2026-04-20T00:07:47Z

## Outcome

Add a chain-scoped guided-session engine that can persist, resume, and move
through planning stages without losing conversational context.

## Why Now

`plan` can run individual shaping passes, but it still behaves like a set of
separate artifact commands. Without durable guided-session state, a true
end-to-end co-planning flow will feel brittle and fragmented.

## Shape

### Appetite

Medium. Build only enough session machinery to make guided planning durable and
predictable.

### Outcome

Users can stop and resume a guided planning chain, see a short summary so far,
and continue from the active stage without reconstructing context by hand.

### Scope Boundary

- one active guided session per planning chain
- repo-level last-active pointer for fast resume
- active stage, stage status, and question-cluster progress
- resume summary plus next-best-action output
- tests for interrupt, resume, and multi-session behavior

### Out of Scope

- multi-user or hosted session state
- cross-repo session coordination
- stage-specific question content
- automatic repo document scanning

### Success Signal

Users can leave a guided planning chain mid-stage, return later, and continue
cleanly with the right summary and stage reopened.

## Scope Boundary

- chain-scoped session model in `.plan/.meta/`
- resume flow in the CLI
- menu state for `continue / refine / stop for now`
- session switching when multiple feature chains exist

Not in scope:

- stage-specific prompt design
- roadmap parking writes
- downstream stale-review logic beyond session metadata hooks

## Spec

- [Draft Spec](../specs/guided-session-engine-and-resume.md)

## Resources

- [Brainstorm](../brainstorms/guided-planning-system.md)
- [Product Direction](../PRODUCT.md)
- [Roadmap](../ROADMAP.md)

## Progress

- Target version: `v8`
- Status: planned

## Notes

This epic should provide the durable spine for the rest of the guided system.
If session state is weak, every later guided stage will feel unreliable.
