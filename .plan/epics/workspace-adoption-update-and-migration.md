---
created_at: "2026-04-16T05:33:06Z"
project: plan
slug: workspace-adoption-update-and-migration
spec: workspace-adoption-update-and-migration
target_version: v2
title: Workspace Adoption, Update, and Migration
type: epic
updated_at: "2026-04-16T05:33:06Z"
---

# Workspace Adoption, Update, and Migration

Created: 2026-04-16T05:33:06Z

## Outcome

Make it safe to adopt `plan` in real repos, evolve the workspace over time, and recover from older layouts or partial setup states.

## Why Now

`v1` can create a fresh workspace, but real adoption requires safer updates, repair flows, and support for repos that are not born clean.

## Scope Boundary

- stronger `plan update`
- repo adoption workflows
- migration inspection and repair
- support for partial or older workspaces

Not in scope:

- giant migration systems
- cloud sync or hosted upgrade logic
- `brain` import behavior beyond future compatibility design

## Spec

- [Draft Spec](../specs/workspace-adoption-update-and-migration.md)

## Resources

- [Product Direction](../PRODUCT.md)
- [Roadmap](../ROADMAP.md)

## Progress

- Target version: `v2`
- Status: planned

## Notes

This epic should make migrations feel like workspace hygiene, not database theater.
