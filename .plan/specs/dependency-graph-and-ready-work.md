---
created_at: "2026-04-16T05:33:06Z"
epic: dependency-graph-and-ready-work
project: plan
slug: dependency-graph-and-ready-work
status: implementing
target_version: v3
title: Dependency Graph and Ready Work Spec
type: spec
updated_at: "2026-04-16T06:57:04Z"
---

# Dependency Graph and Ready Work Spec

Created: 2026-04-16T05:33:06Z

## Why

Users planning bigger systems need help knowing which stories can move now and which are blocked by unfinished work.

## Problem

Without dependency modeling, larger plans become manual sorting exercises and it becomes harder for agents to pick the right next story.

## Goals

- support explicit dependency links
- distinguish ready, blocked, and in-progress work
- keep the model readable and local
- improve next-step selection for users and agents

## Non-Goals

- replacing project management software
- introducing a database-backed task graph
- modeling every possible relationship type in `v3`

## Constraints

- dependency data should fit the markdown-first model
- ready logic must be explainable
- the feature should work well for solo developers before teams

## Solution Shape

- add dependency references between stories and possibly epics
- compute ready work from local plan state
- expose a simple ready-oriented CLI view later
- keep relationship types narrow at first, likely just blockers

## Flows

1. User links work items with dependencies.
2. `plan` computes which stories have no active blockers.
3. User or agent asks for the next ready work.
4. `plan` surfaces the ready set with blocking reasons for the rest.

## Data / Interfaces

- dependency metadata on stories and epics
- future `plan ready` or richer status output
- blocked-reason output for transparency

## Risks / Open Questions

- whether stable IDs become necessary here
- whether dependencies should live in frontmatter or content sections

## Rollout

- start with simple blocker relationships
- avoid advanced graph features until the basic ready model proves useful

## Verification

- blocked stories do not appear in ready views
- ready stories surface with correct rationale
- dependency loops or invalid references are detected clearly

## Story Breakdown

- [ ] [Add dependency metadata to stories](../stories/add-dependency-metadata-to-stories.md)
- [ ] [Compute ready and blocked work sets](../stories/compute-ready-and-blocked-work-sets.md)
- [ ] [Expose ready-work CLI and status views](../stories/expose-ready-work-cli-and-status-views.md)

## Resources

- [Epic](../epics/dependency-graph-and-ready-work.md)
- [Reference Analysis](../research/reference-analysis.md)

## Notes
