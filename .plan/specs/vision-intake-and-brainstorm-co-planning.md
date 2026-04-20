---
created_at: "2026-04-20T00:07:47Z"
epic: vision-intake-and-brainstorm-co-planning
project: plan
slug: vision-intake-and-brainstorm-co-planning
status: approved
target_version: v8
title: Vision Intake and Brainstorm Co-Planning Spec
type: spec
updated_at: "2026-04-20T03:04:50Z"
---

# Vision Intake and Brainstorm Co-Planning Spec

Created: 2026-04-20T00:07:47Z

## Why

Guided planning must start with the user's mental model, not with generated
planning sections.

## Problem

Current brainstorm workflows can create useful notes, but they still risk
feeling like prestructured artifact filling. Users want `plan` to ask about
their vision, gather relevant supporting material they provide, and shape the
idea through back-and-forth before locking it into planning artifacts.

## Goals

- ask the user for raw vision first
- ask the user for relevant docs, links, or research context directly
- shape ideas through 2-4 question clusters
- reflect back once after each cluster
- when a gap appears, explain it and offer one recommended path plus 1-2 alternatives
- challenge vague, bloated, or contradictory thinking while still helping move forward

## Non-Goals

- automatic repo doc discovery
- silent AI-authored planning fields
- end-to-end epic/spec/story creation in this epic

## Constraints

- user-input-first must remain the default
- cluster prompts should stay small and conversational
- reflection should happen once per cluster, not after every answer
- challenge should help, not just criticize
- the brainstorm note remains the durable artifact for this stage

## Solution Shape

- turn brainstorm start into guided vision intake
- ask explicitly for user-supplied docs, links, or research before shaping
- run clustered clarification prompts
- produce a short structured recap with:
  - current understanding
  - key decisions
  - unresolved risks or questions
  - parked items
  - recommended next stage
- keep AI drafting as an explicit opt-in assist

## Flows

1. User starts a brainstorm.
2. `plan` asks for the user's vision in plain language.
3. `plan` asks whether the user has relevant docs, links, or research to provide.
4. `plan` asks a 2-4 question cluster.
5. User answers the cluster.
6. `plan` reflects back once against the whole cluster.
7. If a gap or contradiction appears, `plan` explains it, recommends one path,
   and offers 1-2 alternatives.
8. The stage ends with a recap plus `continue / refine / stop for now`.

## Data / Interfaces

- improved brainstorm-stage command UX
- recap structure for guided brainstorm stage
- gap-handling prompt format
- optional AI-drafting affordance

## Risks / Open Questions

- how to keep challenge useful without making the stage feel adversarial
- how much recap detail is enough before it becomes repetitive

## Rollout

- upgrade brainstorm start and refine first
- layer challenge behavior into the same guided stage
- test partial answers, interruptions, and reruns

## Verification

- brainstorm start asks for vision before structured shaping fields
- supporting docs are requested from the user, not auto-discovered
- clustered prompts stay within 2-4 questions
- reflection occurs once per cluster
- gap handling offers a recommended path plus alternatives
- the stage closes with recap plus `continue / refine / stop for now`

## Story Breakdown

- [ ] [Implement Guided Vision Intake](https://github.com/JimmyMcBride/plan/issues/11)
- [ ] [Add Cluster Reflection And Gap Guidance](https://github.com/JimmyMcBride/plan/issues/14)
- [ ] [Add Brainstorm Stage Recap And Stop Flow](https://github.com/JimmyMcBride/plan/issues/15)

## Analysis

### Missing Constraints

### Success Criteria Gaps

### Hidden Dependencies

### Risk Gaps

### What/Why vs How Leakage

### Recommended Revisions

## Checklist

## Resources

- [Epic](../epics/vision-intake-and-brainstorm-co-planning.md)
- [Brainstorm](../brainstorms/guided-planning-system.md)
- [Product Direction](../PRODUCT.md)

## Notes

Recommendation locked during brainstorming: guided conversation only for now.
No separate fast non-interactive mode.
