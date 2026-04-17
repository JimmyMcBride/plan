---
created_at: "2026-04-17T20:58:34Z"
epic: story-slice-preview-and-apply-flow
project: plan
slug: story-slice-preview-and-apply-flow
status: approved
target_version: v6
title: Story Slice Preview and Apply Flow Spec
type: spec
updated_at: "2026-04-17T20:58:34Z"
---

# Story Slice Preview and Apply Flow Spec

Created: 2026-04-17T20:58:34Z

## Why
`v6` should make the spec-to-story handoff materially better. Right now the
workflow gets to an approved spec, then falls back to manual story creation.

## Problem
Approved specs can describe the right work and still leave maintainers to
hand-copy titles, criteria, and verification into separate story notes. That
invites drift and makes the last planning step feel weaker than the shaping
passes that came before it.

## Goals
- ship `plan story slice`
- preview proposed story notes before writing files
- create todo stories with descriptions, acceptance criteria, verification, and
  canonical spec links
- keep the spec canonical while refreshing `## Story Breakdown` with linked
  story entries after apply

## Non-Goals
- hosted or remote slicing services
- automatic dependency scheduling between stories
- deleting or rewriting existing stories without explicit user action

## Constraints
- only approved specs may be sliced into stories
- preview must work without writing files
- apply must be safe to rerun and avoid duplicate story creation
- manual `plan story create` must remain valid for edge cases

## Solution Shape
- add a candidate slice model with title, description, acceptance criteria, and
  verification
- add `plan story slice <epic-slug>` with preview-first terminal flow and
  explicit apply confirmation
- reuse the existing story template and lifecycle rules when writing notes
- update the spec `## Story Breakdown` with linked story entries after apply

## Flows
1. User approves a spec with concrete implementation guidance.
2. User runs `plan story slice <epic-slug>`.
3. `plan` collects or derives candidate slices and prints a preview.
4. User confirms apply.
5. `plan` creates todo stories and refreshes the spec `## Story Breakdown`
   entries to link to the created notes.

## Data / Interfaces
- `StorySliceInput` or equivalent candidate slice struct in planning code
- `plan story slice <epic-slug>`
- spec `## Story Breakdown` entries that can be rewritten as linked checklists
- story creation helpers reused by slice apply

## Risks / Open Questions
- how much slice data should be parsed from `## Story Breakdown` versus
  collected interactively
- how reruns should behave when some target story slugs already exist
- whether preview should allow partial apply in the first version

## Rollout
- land preview and apply together in `v6`
- keep manual story creation untouched as fallback
- use the command output to inform later readiness checks

## Verification
- preview shows candidate stories without writing files
- apply creates story notes with required sections and canonical spec links
- rerun behavior avoids duplicate story files and preserves deliberate edits
- focused command and planning tests cover first-run and rerun flows

## Story Breakdown
- [ ] [Add story slice candidate model and preview formatting](../stories/add-story-slice-candidate-model-and-preview-formatting.md)
- [ ] [Implement story slice apply flow](../stories/implement-story-slice-apply-flow.md)
- [ ] [Add story slice tests and rerun coverage](../stories/add-story-slice-tests-and-rerun-coverage.md)

## Analysis

### Missing Constraints

### Success Criteria Gaps

### Hidden Dependencies

### Risk Gaps

### What/Why vs How Leakage

### Recommended Revisions

## Checklist

## Resources
- [Epic](../epics/story-slice-preview-and-apply-flow.md)
- [Product Direction](../PRODUCT.md)

## Notes
