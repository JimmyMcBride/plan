---
created_at: "2026-04-16T05:33:06Z"
project: plan
slug: core-workspace-and-artifact-system
spec: core-workspace-and-artifact-system
target_version: v1
title: Core Workspace and Artifact System
type: epic
updated_at: "2026-04-16T05:33:06Z"
---

# Core Workspace and Artifact System

Created: 2026-04-16T05:33:06Z

## Outcome

Deliver a stable `.plan/` workspace that feels trustworthy, obvious, and durable inside any repo.

## Why Now

If the file model is weak, every later planning feature becomes fragile. This epic establishes the permanent ground rules for how `plan` stores and upgrades project planning material.

## Scope Boundary

- `.plan/` directory layout
- `PROJECT.md`, `ROADMAP.md`, `epics/`, `specs/`, `stories/`, `.meta/`
- `plan init`, `plan doctor`, `plan update`
- lightweight workspace metadata and migration state

Not in scope:

- dependency graph
- integrations
- hosted state
- retrieval or memory features

## Spec

- [Draft Spec](../specs/core-workspace-and-artifact-system.md)

## Resources

- [Product Direction](../PRODUCT.md)
- [Roadmap](../ROADMAP.md)

## Progress

- Target version: `v1`
- Status: planned

## Notes

This epic defines the base that all other versions depend on.
