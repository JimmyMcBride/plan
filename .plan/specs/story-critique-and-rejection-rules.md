---
created_at: "2026-04-17T20:58:34Z"
epic: story-critique-and-rejection-rules
project: plan
slug: story-critique-and-rejection-rules
status: done
target_version: v6
title: Story Critique and Rejection Rules Spec
type: spec
updated_at: "2026-04-17T23:36:00Z"
---

# Story Critique and Rejection Rules Spec

Created: 2026-04-17T20:58:34Z

## Why
Story notes now have basic execution-readiness checks, but `v6` should also add
an optional pass that pressures scope and quality before implementation starts.

## Problem
Stories can satisfy the current section-level checks and still be too broad,
missing hidden inputs, or weak enough that they should be rewritten before any
agent or human starts coding.

## Goals
- ship `plan story critique`
- add a durable `## Critique` section to story notes
- record keep, rewrite, or reslice guidance without adding new hierarchy
  objects
- make critique findings useful before a story moves into active execution

## Non-Goals
- multi-reviewer approval workflows
- new story metadata states such as `rejected`
- replacing `plan check` with an interactive critique command

## Constraints
- critique must stay additive to the story note structure
- reruns should update critique safely without damaging the rest of the story
- rejection rules should stay local, explicit, and readable in markdown
- the default flow must still allow direct execution without requiring critique

## Solution Shape
- extend the story template with a durable `## Critique` section
- add `plan story critique <story-slug>` to collect findings about scope fit,
  hidden dependencies, missing contract details, and verification quality
- record a clear recommendation to keep the story, rewrite it, or reslice it
- keep the command grounded in note updates and terminal output rather than new
  workflow objects

## Flows
1. User creates or slices a story from an approved spec.
2. User runs `plan story critique <story-slug>`.
3. `plan` captures critique findings and records the recommended next step.
4. User either keeps the story, rewrites it, or reslices it before starting
   implementation.

## Data / Interfaces
- story template additions under `## Critique`
- critique input and output model in planning code
- `plan story critique <story-slug>`
- note update helpers that preserve the rest of the story body

## Risks / Open Questions
- how to keep critique useful without becoming ceremony
- whether the first version should include explicit scoring or only structured
  findings
- how to phrase rejection guidance clearly without adding a new lifecycle state

## Rollout
- land the note schema and command in `v6`
- keep critique optional but discoverable next to story slicing
- use critique output to inform later `plan check` readiness work

## Verification
- new or existing stories can hold durable critique content
- critique command updates the correct sections without damaging story notes
- critique output clearly distinguishes keep versus rewrite or reslice guidance
- tests cover first-run and rerun critique behavior

## Story Breakdown
- [x] [Add story critique section schema](../stories/add-story-critique-section-schema.md)
- [x] [Implement story critique command flow](../stories/implement-story-critique-command-flow.md)
- [x] [Add story critique tests and docs](../stories/add-story-critique-tests-and-docs.md)

## Analysis

### Missing Constraints

### Success Criteria Gaps

### Hidden Dependencies

### Risk Gaps

### What/Why vs How Leakage

### Recommended Revisions

## Checklist

## Resources
- [Epic](../epics/story-critique-and-rejection-rules.md)
- [Product Direction](../PRODUCT.md)

## Notes
