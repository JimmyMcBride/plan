---
created_at: "2026-04-20T00:07:47Z"
epic: guided-stage-handoffs-and-artifact-writing
project: plan
slug: guided-stage-handoffs-and-artifact-writing
status: approved
target_version: v8
title: Guided Stage Handoffs and Artifact Writing Spec
type: spec
updated_at: "2026-04-20T03:04:50Z"
---

# Guided Stage Handoffs and Artifact Writing Spec

Created: 2026-04-20T00:07:47Z

## Why

Guided planning only becomes real product leverage if it carries the user from
one stage to the next instead of making them manually reassemble the workflow.

## Problem

The current product has good artifact types and shaping passes, but the user
still has to decide when to switch modes, which command to run next, and how to
translate one stage into the next. That keeps `plan` too tool-like and not
guide-like.

## Goals

- improve existing commands into one coherent stage-by-stage flow
- guide users from brainstorm to epic to spec to stories
- write durable artifacts at clear stage checkpoints
- end each stage with a recap plus numbered `continue / refine / stop for now`
- keep story outputs execution-ready and verification-aware

## Non-Goals

- a separate top-level guided command family
- replacing the canonical artifact model
- portfolio or roadmap planning outside the current feature chain

## Constraints

- build on top of current brainstorm/epic/spec/story commands
- keep the flow stage-based rather than one giant wizard
- do not force AI drafting when the user wants to author the plan
- recap shape should stay consistent across stages

## Solution Shape

- improve current stage commands into guided checkpoints
- use a shared recap contract:
  - current understanding
  - key decisions
  - unresolved risks or questions
  - parked items
  - recommended next stage
- use numbered stage menus for `continue / refine / stop for now`
- write or update artifacts when a stage checkpoint closes
- carry forward enough context to avoid the user restating shaped decisions

## Flows

1. User completes guided brainstorm stage.
2. `plan` shows recap and recommends moving to epic.
3. User chooses `continue`.
4. `plan` creates or updates the epic, then runs the guided epic stage.
5. The same pattern repeats for spec and then story creation.
6. At any stage, the user can choose `refine` or `stop for now`.
7. If the user stops, `plan` saves state and prints the next-best action.

## Data / Interfaces

- guided stage menus in current commands
- stage recap contract
- artifact write points for epic/spec/story stages
- session summary output for stop/resume

## Risks / Open Questions

- how much artifact writing should happen before versus after user confirmation
- how to keep the story stage guided without making it feel ceremonial

## Rollout

- integrate brainstorm-to-epic handoff first
- add epic-to-spec next
- finish with guided story creation and next-action summaries

## Verification

- users can move from brainstorm to epic to spec to stories without leaving the guided flow
- each stage ends with recap plus `continue / refine / stop for now`
- artifact writes happen at stable checkpoints rather than every partial answer
- stopping mid-flow preserves a clear next-best-action summary

## Story Breakdown

## Analysis

### Missing Constraints

### Success Criteria Gaps

### Hidden Dependencies

### Risk Gaps

### What/Why vs How Leakage

### Recommended Revisions

## Checklist

## Resources

- [Epic](../epics/guided-stage-handoffs-and-artifact-writing.md)
- [Brainstorm](../brainstorms/guided-planning-system.md)
- [Product Direction](../PRODUCT.md)

## Notes

Recommendation locked during brainstorming: improve and replace the current
stage commands under the hood rather than inventing a separate guided family.
