---
created_at: "2026-04-17T20:19:39Z"
project: plan
slug: spec-checklist-profiles
spec: spec-checklist-profiles
title: Spec Checklist Profiles
type: epic
updated_at: "2026-04-17T20:19:39Z"
---

# Spec Checklist Profiles

Created: 2026-04-17T20:19:39Z

## Outcome

Add reusable spec checklist profiles that catch domain-specific planning gaps
without rewriting the canonical spec sections.

## Why Now

`plan spec analyze` gives one general diagnostic pass. `v5` needs a second,
profile-driven pass that helps users review UI flows, integrations, and
migrations with more tailored questions.

## Scope Boundary

- `## Checklist` support in spec notes
- a `plan spec checklist` command with named profiles
- blocking versus advisory findings where appropriate
- docs and tests for stable checklist behavior

Not in scope:

- code generation
- organization-specific checklist registries
- auto-approval of specs

## Spec

- [Draft Spec](../specs/spec-checklist-profiles.md)

## Resources

- [Product Direction](../PRODUCT.md)
- [Roadmap](../ROADMAP.md)

## Progress

- Target version: `v5`
- Status: planned

## Notes

Checklist profiles should deepen planning rigor only when users ask for the
extra pass.
