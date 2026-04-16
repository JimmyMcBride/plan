---
created_at: "2026-04-16T05:33:06Z"
epic: spec-driven-planning-workflow
project: plan
slug: spec-driven-planning-workflow
status: draft
target_version: v1
title: Spec-Driven Planning Workflow Spec
type: spec
updated_at: "2026-04-16T05:33:06Z"
---

# Spec-Driven Planning Workflow Spec

Created: 2026-04-16T05:33:06Z

## Why

The whole product promise depends on turning rough ideas into execution-ready stories through a clean spec gate.

## Problem

Developers need a planning flow that is simple enough to use constantly and strict enough to stop vague, low-quality work from reaching execution.

## Goals

- support brainstorm creation and idea capture
- promote brainstorms into epics
- create one canonical spec per epic
- require explicit spec approval before stories
- track stories and project status cleanly

## Non-Goals

- generating code
- memory or search workflows
- adding tasks below stories as first-class objects

## Constraints

- keep the hierarchy `Epic -> Spec -> Story`
- treat brainstorms as workflow entry, not canonical hierarchy
- stories must be small enough to execute in one focused pass

## Solution Shape

- brainstorms collect ideas and questions
- epics define outcome and scope boundary
- specs become the canonical contract
- stories become execution-ready units with acceptance and verification

## Flows

1. Start brainstorm.
2. Add ideas.
3. Promote brainstorm to epic and seeded draft spec.
4. Refine and approve spec.
5. Create stories from the approved spec.
6. Track execution progress with story status and overall `plan status`.

## Data / Interfaces

- brainstorm frontmatter tracks slug and status
- epic frontmatter links canonical spec slug
- spec frontmatter tracks approval status
- story frontmatter tracks epic, spec, and execution status

## Risks / Open Questions

- how much story decomposition help should exist in v1 versus v2
- whether status transitions need stricter validation later

## Rollout

- ship basic end-to-end flow in `v1`
- keep spec approval lightweight
- defer plan quality engine to `v2`

## Verification

- user can complete the full brainstorm -> epic -> spec -> story flow locally
- story creation fails when spec is still draft
- status output reflects epic/story progress accurately

## Story Breakdown

- [ ] Finalize brainstorm note behavior
- [ ] Harden epic promotion and spec seeding
- [ ] Enforce spec approval gate for story creation
- [ ] Improve status reporting across epics and stories

## Resources

- [Epic](../epics/spec-driven-planning-workflow.md)
- [Product Direction](../PRODUCT.md)

## Notes
