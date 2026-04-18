---
created_at: "2026-04-17T20:58:34Z"
project: plan
slug: story-slice-preview-and-apply-flow
spec: story-slice-preview-and-apply-flow
title: Story Slice Preview and Apply Flow
type: epic
updated_at: "2026-04-17T20:58:34Z"
---

# Story Slice Preview and Apply Flow

Created: 2026-04-17T20:58:34Z

## Outcome
Give approved specs a durable path to turn story ideas into execution-ready
story notes with a preview step before files are written.

## Why Now
The workflow gets strong up through spec approval, then drops back to manual
copying when stories need to be created. `v6` should make that handoff feel as
deliberate as the earlier shaping passes.

## Shape

### Appetite
One small `v6` pass: approved spec in, candidate slices previewed, selected
stories written safely to `.plan/stories/`.

### Outcome
Users can inspect the first-pass story set before apply, then create linked
todo stories without hand-copying titles, criteria, and verification steps.

### Scope Boundary
- `plan story slice <epic-slug>`
- preview-first terminal flow for candidate stories
- apply flow that writes story notes and refreshes spec story links
- slice data shaped around title, description, acceptance criteria, and
  verification

### Out of Scope
- hosted AI slicing services
- dependency graph generation
- removing manual `plan story create`

### Success Signal
An approved spec can be turned into a clean first-pass story backlog in one
local loop without losing the canonical spec contract.

## Scope Boundary
- approved spec to story transition
- preview and apply UX
- story note creation helpers
- spec Story Breakdown updates after apply

Not in scope:

- story critique rules
- broader project orchestration
- external tracker sync

## Spec
- [Draft Spec](../specs/story-slice-preview-and-apply-flow.md)

## Resources
- [Product Direction](../PRODUCT.md)
- [Roadmap](../ROADMAP.md)

## Progress
- Target version: `v6`
- Status: planned

## Notes
The preview matters. `v6` should help maintainers inspect the slice before the
repo gains new story files.
