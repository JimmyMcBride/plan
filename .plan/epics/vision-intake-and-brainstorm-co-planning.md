---
created_at: "2026-04-20T00:07:47Z"
project: plan
slug: vision-intake-and-brainstorm-co-planning
spec: vision-intake-and-brainstorm-co-planning
title: Vision Intake and Brainstorm Co-Planning
type: epic
updated_at: "2026-04-20T00:07:47Z"
---

# Vision Intake and Brainstorm Co-Planning

Created: 2026-04-20T00:07:47Z

## Outcome

Turn brainstorm start and early shaping into a real co-planning conversation
that begins with user vision and user-supplied context.

## Why Now

The first experience with `plan` still feels too much like artifact creation.
If the opening stage is not collaborative, the rest of the guided system will
feel like a nicer shell on the same old workflow.

## Shape

### Appetite

Medium. Focus on the first guided stage only: vision, context, clarification,
and challenge.

### Outcome

Users can start with raw vision, share relevant docs or research, answer small
question clusters, and get a reflected understanding before any structured
artifact hardens.

### Scope Boundary

- ask users directly for vision plus relevant docs, links, or research
- 2-4 question clusters for early shaping
- one reflection pass after each cluster
- opinionated gap handling with one recommended path plus 1-2 alternatives
- challenge behavior for vagueness, bloat, and contradiction

### Out of Scope

- epic, spec, and story writing beyond the brainstorm stage
- auto-scanning repo docs
- silent AI drafting without user buy-in

### Success Signal

Starting a brainstorm feels like collaborative thinking, not form filling.

## Scope Boundary

- improved `brainstorm start`, `brainstorm refine`, and `brainstorm challenge`
- direct ask for user-supplied supporting material
- recap content for the brainstorm stage

Not in scope:

- cross-stage handoff behavior
- stale downstream review
- roadmap parking writes beyond suggestion points

## Spec

- [Draft Spec](../specs/vision-intake-and-brainstorm-co-planning.md)

## Resources

- [Brainstorm](../brainstorms/guided-planning-system.md)
- [Product Direction](../PRODUCT.md)
- [Roadmap](../ROADMAP.md)

## Progress

- Target version: `v8`
- Status: planned

## Notes

This epic should establish the default feel of guided planning. If this stage
feels wrong, later stage guidance will not save the experience.
