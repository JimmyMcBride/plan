---
created_at: "2026-04-17T20:19:38Z"
project: plan
slug: model-aware-planning-skills
spec: model-aware-planning-skills
title: Model-Aware Planning Skills
type: epic
updated_at: "2026-04-17T20:19:38Z"
---

# Model-Aware Planning Skills

Created: 2026-04-17T20:19:38Z

## Outcome

Ship a stronger installed `plan` skill bundle that steers GPT-style and
reasoning-heavy models differently while preserving the same local-first
planning contract.

## Why Now

The repo now has the first refinement and analysis passes, but the installed
skill still behaves like a thin generic wrapper. `v5` should make planning
guidance itself a product surface instead of relying on note templates alone.

## Scope Boundary

- model-family guidance in the bundled skill
- explicit planning passes and completion rules
- examples and rubrics that match `plan`'s artifact model
- verification that the shipped skill text matches the new product story

Not in scope:

- hosted prompt orchestration
- model routing outside the skill bundle
- cloud memory or context systems

## Spec

- [Draft Spec](../specs/model-aware-planning-skills.md)

## Resources

- [Product Direction](../PRODUCT.md)
- [Roadmap](../ROADMAP.md)

## Progress

- Target version: `v5`
- Status: planned

## Notes

This epic turns prompt quality into a first-class part of the product rather
than an afterthought bolted onto markdown templates.
