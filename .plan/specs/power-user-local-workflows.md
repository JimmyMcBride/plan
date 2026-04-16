---
created_at: "2026-04-16T05:33:06Z"
epic: power-user-local-workflows
project: plan
slug: power-user-local-workflows
status: implementing
target_version: v3
title: Power-User Local Workflows Spec
type: spec
updated_at: "2026-04-16T07:12:42Z"
---

# Power-User Local Workflows Spec

Created: 2026-04-16T05:33:06Z

## Why

The best local tools scale upward without forcing everyone into advanced mode. `plan` needs a path for power users that still respects the simple default flow.

## Problem

As projects grow, users need stronger local conventions around multiple epics, versions, branches, and deeper workflow visibility.

## Goals

- support advanced local planning workflows
- improve visibility across larger bodies of work
- keep the base model intact while adding optional power
- preserve indie-dev friendliness even as complexity grows

## Non-Goals

- simulating enterprise PM systems
- hosted team coordination
- replacing external integrations that are explicitly deferred

## Constraints

- advanced workflows must be optional
- the simple path must still feel first-class
- features should work repo-locally without central services

## Solution Shape

- add richer views and helpers around roadmap, status, and dependencies
- support stronger local conventions for bigger repos
- leave external integrations in the parking lot until after local workflows are excellent

## Flows

1. User manages multiple epics and versions locally.
2. `plan` surfaces clearer views of what is active, blocked, and upcoming.
3. User keeps planning local even as project size grows.

## Data / Interfaces

- richer local summary views
- future branch-aware or workspace-aware helpers
- stronger aggregation across roadmap, epics, specs, and stories

## Risks / Open Questions

- where helpful workflow guidance ends and ceremony begins
- whether stable IDs become necessary for branch-heavy usage

## Rollout

- prioritize views and helpers that reinforce the existing model
- defer anything that smells like external PM replication

## Verification

- advanced users can manage larger plans without leaving the repo
- new users can still ignore advanced surfaces and succeed
- local-first remains true even as workflows deepen

## Story Breakdown

- [ ] [Add filtered status views for large local plans](../stories/add-filtered-status-views-for-large-local-plans.md)
- [ ] [Add advanced roadmap and story selection helpers](../stories/add-advanced-roadmap-and-story-selection-helpers.md)
- [ ] [Document and test advanced local planning workflows](../stories/document-and-test-advanced-local-planning-workflows.md)

## Resources

- [Epic](../epics/power-user-local-workflows.md)
- [Roadmap](../ROADMAP.md)

## Notes
