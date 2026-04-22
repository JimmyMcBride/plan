---
name: plan
description: Use this skill when working with a project-local `plan` workspace that keeps planning material under `.plan/`. Focus on brainstorms, idea docs, specs, execution slices, legacy compatibility surfaces, and roadmap updates.
user-invocable: true
args:
  - name: task
    description: The planning task to perform.
    required: false
---

# Plan

Use `plan` as the primary interface for project planning, while respecting the
configured source-of-truth backend.

## Goals

- keep local planning as the default while respecting `github` and `hybrid`
  source-of-truth modes when configured
- treat specs as the canonical execution contract
- improve the quality of brainstorms and specs before implementation starts
- use lightweight initiative metadata when multiple specs belong together
- guide execution from approved specs without persisting tiny slice artifacts by default
- avoid sidecar planning systems outside `.plan/`

## Startup

When a repo uses `plan`:

1. Read `.plan/PROJECT.md`.
2. Read `.plan/ROADMAP.md`.
3. Read the brainstorm, idea, spec, and legacy compatibility notes relevant to the task.
4. If the project is using `github` or `hybrid` ownership for durable planning data, inspect the linked GitHub issue, project, or milestone state too.
5. Use the `plan` CLI for durable planning changes and keep backend ownership explicit.

## Commands

- `plan init --project .`
- `plan adopt --project .`
- `plan doctor --project .`
- `plan update --project .`
- `plan update --project . --archive-legacy`
- `plan check --project .`
- `plan brainstorm start --project . "<topic>"`
- `plan brainstorm idea --project . <brainstorm-slug> --body "<idea>"`
- `plan brainstorm refine --project . <brainstorm-slug>`
- `plan brainstorm challenge --project . <brainstorm-slug>`
- `plan epic create|promote|shape ...` only when a repo still depends on the legacy transition path
- `plan spec show --project . <spec-slug>`
- `plan spec analyze --project . <spec-slug>`
- `plan spec checklist --project . <spec-slug> --profile general`
- `plan spec status --project . <spec-slug> --set approved`
- `plan spec initiative --project . <spec-slug> --set <initiative-slug>`
- `plan spec execute --project . <spec-slug>`
- `plan story critique --project . <story-slug>`
- `plan story create|update|slice ...` only for legacy compatibility during migration
- `plan roadmap show --project .`
- `plan status --project .`

## Rules

- Brainstorms are discovery material, not the canonical hierarchy.
- Active model is `Brainstorm -> Idea Doc (optional) -> Spec`, with runtime slices during execution.
- Local is the default backend, not the only backend.
- Do not assume every durable planning artifact lives in `.plan/`; respect explicit ownership by planning layer.
- `brainstorm refine` should reduce ambiguity before promotion.
- `brainstorm challenge` should pressure-test risk, no-gos, and overengineering before promotion.
- `epic shape` is now a legacy compatibility pass, not the preferred active model.
- `spec analyze` should pressure-test a spec without rewriting its canonical sections.
- `spec checklist` should add profile-driven rigor without mutating the canonical sections.
- `spec execute` should derive ephemeral execution slices from the canonical spec and suggest a branch-per-spec path.
- `story critique` should reject broad or verification-thin stories before implementation starts.
- Keep roadmap guidance lightweight.
- Do not add a new heavyweight planning layer just to replace epics.
- Keep planning separate from memory, retrieval, or context management systems.

## Planning Modes

Use the smallest pass that resolves the current planning gap:

1. `brainstorm refine` for ambiguity reduction
2. `brainstorm challenge` for rabbit holes, no-gos, and simplification pressure
3. `epic shape` only when the repo is still using the legacy transition path
4. `spec analyze` for general refinement gaps
5. `spec checklist` for domain-specific review
6. `spec execute` for starting branch-per-spec execution from an approved spec
7. `story critique` only for legacy story flows that still need execution-readiness pressure

## Model Guidance

### GPT-style Models

- prefer explicit step order
- restate the artifact you are editing before making changes
- keep output contracts concrete and named
- ask clarifying questions only when ambiguity would materially damage the plan
- use the lightest shaping pass that can resolve the gap

### Reasoning-Heavy Models

- start from the product goal and current artifact quality
- search for second-order issues such as missing non-goals, hidden dependencies, and rollout gaps
- keep recommendations bounded; do not sprawl into new systems
- verify the artifact remains simple after adding rigor

## Completion Contract

- specs stay canonical
- shaping passes stay additive
- optional rigor must not make the default path ceremonial
- every recommendation should improve clarity, boundedness, verification, or executability
- active execution should stay traceable from spec -> slices -> commits -> PR
- when backend ownership is split, the agent should preserve that split instead of silently moving truth back into `.plan/`

## Ambiguity Handling

- if the next shaping pass is obvious, run it
- if two passes could apply, choose the lighter one first
- if backend ownership is unclear, inspect project rules and current GitHub state before editing durable planning artifacts
- do not turn `plan` into memory, context, or execution orchestration
