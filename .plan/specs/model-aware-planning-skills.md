---
created_at: "2026-04-17T20:19:38Z"
epic: model-aware-planning-skills
project: plan
slug: model-aware-planning-skills
status: done
target_version: v5
title: Model-Aware Planning Skills Spec
type: spec
updated_at: "2026-04-17T20:34:01Z"
---

# Model-Aware Planning Skills Spec

Created: 2026-04-17T20:19:38Z

## Why

The skill bundle is the product's main steering surface for agent behavior. It
should encode the new planning thesis explicitly instead of assuming every model
needs the same prompt shape.

## Problem

`plan` now ships refinement and analysis passes, but the installed skill does
not help models apply them consistently or differently based on model style.

## Goals

- add model-aware planning guidance to the bundled skill
- make GPT-style guidance more explicit, step-ordered, and example-driven
- make reasoning-model guidance higher-level but still rubric-aware
- align the skill text with the v4-v7 product reset

## Non-Goals

- dynamic prompt generation services
- cloud profile management
- replacing local markdown artifacts with prompt-only workflows

## Constraints

- the skill must stay local-first and repo-friendly
- guidance must preserve `Brainstorm -> Epic -> Spec -> Story`
- the simple default path must remain discoverable
- installed skill files should be easy to inspect and diff

## Solution Shape

- expand `skills/plan/SKILL.md` with explicit planning modes and model guidance
- add richer agent guidance files for the installed bundle
- include completion rules, verification expectations, and ambiguity behavior
- keep the bundle static and versioned with the repo

## Flows

1. Maintainer updates the skill text and agent guidance files.
2. User installs the skill with `plan skills install`.
3. The installed bundle gives models sharper planning instructions that match
   the current roadmap and passes.
4. Maintainers verify the shipped bundle matches repo expectations.

## Data / Interfaces

- `skills/plan/SKILL.md`
- `skills/plan/agents/*.yaml`
- install targets under Codex and OpenClaw skill roots

## Risks / Open Questions

- how much variation between model families is helpful before the bundle becomes
  too complex
- whether the current agent manifest format needs more structure than YAML text

## Rollout

- ship updated skill text and agent guidance in `v5`
- verify installation behavior against existing targets
- expand to more providers only if the local bundle remains easy to reason about

## Verification

- the repo skill text documents GPT-style and reasoning-model guidance
- installed bundle files include the new guidance
- tests or golden checks fail if the shipped bundle loses the new structure

## Story Breakdown

- [x] [Expand plan skill with model-aware guidance](../stories/expand-plan-skill-with-model-aware-guidance.md)
- [x] [Add model-family agent guidance files](../stories/add-model-family-agent-guidance-files.md)
- [x] [Validate installed skill bundle behavior](../stories/validate-installed-skill-bundle-behavior.md)

## Analysis

### Missing Constraints

### Success Criteria Gaps

### Hidden Dependencies

### Risk Gaps

### What/Why vs How Leakage

### Recommended Revisions

## Resources

- [Epic](../epics/model-aware-planning-skills.md)
- [Product Direction](../PRODUCT.md)

## Notes
