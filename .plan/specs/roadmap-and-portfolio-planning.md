---
created_at: "2026-04-16T05:33:06Z"
epic: roadmap-and-portfolio-planning
project: plan
slug: roadmap-and-portfolio-planning
status: approved
target_version: v2
title: Roadmap and Portfolio Planning Spec
type: spec
updated_at: "2026-04-16T06:28:46Z"
---

# Roadmap and Portfolio Planning Spec

Created: 2026-04-16T05:33:06Z

## Why

Users need a version-level view above epics, but they do not need phase bureaucracy.

## Problem

The core workflow handles individual pieces of work well, but there is not yet a strong portfolio layer that answers, "what is in v1, what moves to v2, and what is parked?"

## Goals

- make `ROADMAP.md` a first-class artifact
- support version sections and version summaries
- show ordered epics per version
- keep parking-lot work visible without promoting it too early

## Non-Goals

- creating a sprint or phase system
- turning roadmap planning into issue tracking
- building dependency execution logic

## Constraints

- roadmap must stay readable in plain markdown
- version planning must remain lightweight
- roadmap should reinforce epics, not replace them

## Solution Shape

- structure `ROADMAP.md` around versions and summaries
- map epics to target versions
- surface parked work and deferred integrations explicitly
- add CLI helpers later for common roadmap edits and views

## Flows

1. Define version goals.
2. Assign epics to versions.
3. Keep ordering notes and parking lot current.
4. Read roadmap summaries before selecting the next epic to detail or execute.

## Data / Interfaces

- version sections in `ROADMAP.md`
- epic frontmatter field for target version
- future CLI helpers for roadmap read/edit flows

## Risks / Open Questions

- how much roadmap automation is useful before it becomes ceremony
- whether version summaries belong only in roadmap or also in generated views

## Rollout

- establish the roadmap format in `v2`
- add richer roadmap commands after the markdown shape is proven

## Verification

- users can summarize `v1`, `v2`, and `v3` directly from `ROADMAP.md`
- epics clearly map to versions
- roadmap stays concise and readable in raw markdown

## Story Breakdown

- [ ] [Define roadmap version structure and parsing](../stories/define-roadmap-version-structure-and-parsing.md)
- [ ] [Add roadmap CLI helpers for version views](../stories/add-roadmap-cli-helpers-for-version-views.md)
- [ ] [Surface roadmap progress in status output](../stories/surface-roadmap-progress-in-status-output.md)

## Resources

- [Epic](../epics/roadmap-and-portfolio-planning.md)
- [Roadmap](../ROADMAP.md)

## Notes
