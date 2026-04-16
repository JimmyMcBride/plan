---
created_at: "2026-04-16T05:33:06Z"
project: plan
slug: dependency-graph-and-ready-work
spec: dependency-graph-and-ready-work
target_version: v3
title: Dependency Graph and Ready Work
type: epic
updated_at: "2026-04-16T05:33:06Z"
---

# Dependency Graph and Ready Work

Created: 2026-04-16T05:33:06Z

## Outcome

Model story and epic dependencies so `plan` can tell the user what is blocked and what is ready to execute next.

## Why Now

After the core model and quality layer are stable, dependency awareness is the first major power feature that can make `plan` dramatically more useful on larger projects.

## Scope Boundary

- dependency links between planning objects
- blocked versus ready views
- local-first dependency reasoning

Not in scope:

- distributed task database
- issue-tracker replacement
- cloud coordination

## Spec

- [Draft Spec](../specs/dependency-graph-and-ready-work.md)

## Resources

- [Reference Analysis](../research/reference-analysis.md)
- [Roadmap](../ROADMAP.md)

## Progress

- Target version: `v3`
- Status: planned

## Notes

This is the first epic that intentionally borrows power-user ideas from `beads`.
