---
created_at: "2026-04-17T20:58:34Z"
project: plan
slug: story-critique-and-rejection-rules
spec: story-critique-and-rejection-rules
title: Story Critique and Rejection Rules
type: epic
updated_at: "2026-04-17T20:58:34Z"
---

# Story Critique and Rejection Rules

Created: 2026-04-17T20:58:34Z

## Outcome
Add an optional critique pass that pressures stories before implementation and
makes rewrite or reslice decisions durable.

## Why Now
Current story checks enforce section presence, but they do not catch stories
that are still too broad, unclear, or risky to start. `v6` should tighten that
execution boundary.

## Shape

### Appetite
One local critique loop that records findings inside the story note and helps a
maintainer decide whether the story should proceed, be rewritten, or be split.

### Outcome
Stories gain a durable critique section and a repeatable command for finding
scope leaks, missing inputs, and weak verification before work starts.

### Scope Boundary
- `plan story critique <story-slug>`
- durable critique content stored on the story note
- explicit keep, rewrite, or reslice guidance
- rejection rules grounded in execution readiness, not reviewer bureaucracy

### Out of Scope
- multi-reviewer approval workflows
- new permanent story lifecycle states such as `rejected`
- external issue-tracker integrations

### Success Signal
Maintainers can run one critique pass and know whether a story is ready to
start or needs another slicing/editing pass first.

## Scope Boundary
- story note schema additions
- critique command flow
- rejection heuristics for too-broad or under-specified stories

Not in scope:

- mandatory critique before every story update
- automated story dependencies
- remote policy engines

## Spec
- [Draft Spec](../specs/story-critique-and-rejection-rules.md)

## Resources
- [Product Direction](../PRODUCT.md)
- [Roadmap](../ROADMAP.md)

## Progress
- Target version: `v6`
- Status: planned

## Notes
Critique should sharpen stories, not turn `plan` into a review queue tool.
