---
created_at: "2026-04-21T06:36:20Z"
project: plan
slug: guide-packet-and-cli-foundation
spec: guide-packet-and-cli-foundation
title: Guide Packet and CLI Foundation
type: epic
updated_at: "2026-04-21T06:36:20Z"
---

# Guide Packet and CLI Foundation

Created: 2026-04-21T06:36:20Z

## Outcome

Introduce a runtime guide-packet interface so agents can ask `plan` for the
current planning mode and stage contract instead of relying on static installed
persona text.

## Why Now

The guided system is gaining durable sessions, stage checkpoints, and handoff
behavior, but `plan` still has no first-class way to hand live guidance to an
external agent. Without that contract, stage quality still depends on stale
skill text and repo-specific prompt glue.

## Shape

### Appetite

Small-to-medium. Ship one stable packet contract, one stage family, and two CLI
entry points before expanding to later stages or richer renderers.

### Outcome

An agent can fetch a live brainstorm-stage guide packet with the current
artifact, session summary, stage checkpoint, next action, and a structured
behavior contract that is safe to follow directly.

### Scope Boundary

- machine-first guide packet `v1`
- packet builder sourced from guided session state, workspace rules, and the
  linked brainstorm note
- `plan guide current`
- `plan guide show`
- brainstorm-stage checkpoints only
- JSON output only
- rendered prompt text derived from the structured packet

### Out of Scope

- direct model API calls from `plan`
- per-agent installed personas as the primary runtime interface
- markdown or plain-text guide renderers
- epic, spec, or story-stage packets
- automatic runtime or skill-consumption wiring

### Success Signal

A generic agent can call `plan` and receive enough live guidance to run the
brainstorm stage cleanly without bespoke repo-local prompt glue.

## Scope Boundary

- reuse existing guided session state instead of adding parallel state
- keep `plan` planning-only
- keep the command surface to `current` and `show` for the first slice
- keep the packet stable and versioned before layering on richer outputs

Not in scope:

- schema export command in the first slice
- family overlays beyond the default packet behavior
- bootstrap skill rewiring in the same epic

## Spec

- [Draft Spec](../specs/guide-packet-and-cli-foundation.md)

## Resources

- [Research: Guide Packet Schema and CLI Design](../research/guide-packet-schema-and-cli-design.md)
- [Guided Session Engine and Resume Spec](../specs/guided-session-engine-and-resume.md)
- [Vision Intake and Brainstorm Co-Planning Spec](../specs/vision-intake-and-brainstorm-co-planning.md)
- [Guided Stage Handoffs and Artifact Writing Spec](../specs/guided-stage-handoffs-and-artifact-writing.md)

## Progress

- Target version: `v8`
- Status: planned

## Notes

This epic should establish `plan` as the runtime source of truth for guided
agent behavior without turning it into a model runner or orchestration layer.
