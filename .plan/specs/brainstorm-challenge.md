---
created_at: "2026-04-17T20:19:38Z"
epic: brainstorm-challenge
project: plan
slug: brainstorm-challenge
status: done
target_version: v5
title: Brainstorm Challenge Spec
type: spec
updated_at: "2026-04-17T20:34:02Z"
---

# Brainstorm Challenge Spec

Created: 2026-04-17T20:19:38Z

## Why

Refinement makes brainstorms clearer, but it does not yet challenge them.

## Problem

Users can still promote brainstorms into epics without a durable pass that
captures rabbit holes, no-gos, assumptions, or simpler alternatives.

## Goals

- add a durable `## Challenge` section to brainstorm notes
- ship `plan brainstorm challenge`
- keep challenge passes resumable and idempotent
- make the output useful before epic promotion

## Non-Goals

- automatic rejection of brainstorms
- remote AI-provider integrations
- replacing the lighter-weight brainstorm idea capture flow

## Constraints

- brainstorm notes remain the only durable artifact for this pass
- challenge updates should not rewrite canonical brainstorm content
- the command should work well in an interactive TTY loop
- reruns should resume from what is still missing

## Solution Shape

- extend the brainstorm template with a fixed `## Challenge` section
- add a command that collects challenge data in clustered prompts
- persist after each cluster using the existing note update primitives
- keep the output simple markdown with no extra workflow state

## Flows

1. User starts or refines a brainstorm.
2. User runs `plan brainstorm challenge <brainstorm-slug>`.
3. `plan` prompts for rabbit holes, no-gos, assumptions, overengineering, and a
   simpler alternative.
4. The brainstorm note is updated after each cluster and can be resumed later.

## Data / Interfaces

- brainstorm template additions
- a `ChallengeInput` shape in planning code
- the `plan brainstorm challenge` command

## Risks / Open Questions

- how to keep the prompt loop useful without becoming too ceremonial
- whether some sections should allow empty answers without blocking completion

## Rollout

- land the note schema and command together
- verify reruns and partial updates in tests
- keep the pass optional but easy to discover next to `refine`

## Verification

- new brainstorms include the `## Challenge` headings
- `plan brainstorm challenge` updates only the challenge section
- reruns preserve earlier answers and prompt only for missing fields where
  possible

## Story Breakdown

- [x] [Add brainstorm challenge note schema](../stories/add-brainstorm-challenge-note-schema.md)
- [x] [Implement brainstorm challenge command flow](../stories/implement-brainstorm-challenge-command-flow.md)
- [x] [Add brainstorm challenge tests and resume coverage](../stories/add-brainstorm-challenge-tests-and-resume-coverage.md)

## Analysis

### Missing Constraints

### Success Criteria Gaps

### Hidden Dependencies

### Risk Gaps

### What/Why vs How Leakage

### Recommended Revisions

## Resources

- [Epic](../epics/brainstorm-challenge.md)
- [Product Direction](../PRODUCT.md)

## Notes
