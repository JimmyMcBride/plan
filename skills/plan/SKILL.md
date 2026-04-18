---
name: plan
description: Use this skill when working with a project-local `plan` workspace that keeps planning material under `.plan/`. Focus on brainstorms, refinement, epics, specs, stories, and roadmap updates.
user-invocable: true
args:
  - name: task
    description: The planning task to perform.
    required: false
---

# Plan

Use `plan` as the primary interface for repo-local planning.

## Goals

- keep planning local to the repo
- treat specs as the canonical execution contract
- improve the quality of brainstorms and specs before implementation starts
- create stories only after spec approval
- keep stories execution-ready and verification-aware
- avoid sidecar planning systems outside `.plan/`

## Startup

When a repo uses `plan`:

1. Read `.plan/PROJECT.md`.
2. Read `.plan/ROADMAP.md`.
3. Read the brainstorm, epic, spec, and story notes relevant to the task.
4. Use the `plan` CLI for durable planning changes.

## Commands

- `plan init --project .`
- `plan adopt --project .`
- `plan doctor --project .`
- `plan update --project .`
- `plan check --project .`
- `plan brainstorm start --project . "<topic>"`
- `plan brainstorm idea --project . <brainstorm-slug> --body "<idea>"`
- `plan brainstorm refine --project . <brainstorm-slug>`
- `plan brainstorm challenge --project . <brainstorm-slug>`
- `plan epic create --project . "<title>"`
- `plan epic promote --project . <brainstorm-slug>`
- `plan epic shape --project . <epic-slug>`
- `plan spec show --project . <epic-slug>`
- `plan spec analyze --project . <epic-slug>`
- `plan spec checklist --project . <epic-slug> --profile general`
- `plan spec status --project . <epic-slug> --set approved`
- `plan story slice --project . <epic-slug>`
- `plan story critique --project . <story-slug>`
- `plan story create --project . <epic-slug> "<title>" --criteria "<criterion>" --verify "<step>"`
- `plan story update --project . <story-slug> --status in_progress`
- `plan roadmap show --project .`
- `plan status --project .`

## Rules

- Brainstorms are discovery material, not the canonical hierarchy.
- Canonical hierarchy is `Epic -> Spec -> Story`.
- `brainstorm refine` should reduce ambiguity before promotion.
- `brainstorm challenge` should pressure-test risk, no-gos, and overengineering before promotion.
- `epic shape` should turn an epic into a bounded bet with appetite and success signal.
- `spec analyze` should pressure-test a spec without rewriting its canonical sections.
- `spec checklist` should add profile-driven rigor without mutating the canonical sections.
- `story slice` should stay preview-first and derive execution-ready slices from the canonical spec.
- `story critique` should reject broad or verification-thin stories before implementation starts.
- Keep roadmap guidance lightweight.
- Do not add tasks beneath stories as first-class objects unless the project explicitly asks for that system.
- Keep planning separate from memory, retrieval, or context management systems.

## Planning Modes

Use the smallest pass that resolves the current planning gap:

1. `brainstorm refine` for ambiguity reduction
2. `brainstorm challenge` for rabbit holes, no-gos, and simplification pressure
3. `epic shape` for appetite and scope boundaries
4. `spec analyze` for general refinement gaps
5. `spec checklist` for domain-specific review
6. `story slice` for turning approved spec breakdowns into first-pass story sets
7. `story critique` for execution-readiness pressure before coding

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
- spec-to-story handoffs should stay checkable with `plan check`

## Ambiguity Handling

- if the next shaping pass is obvious, run it
- if two passes could apply, choose the lighter one first
- do not turn `plan` into memory, context, or execution orchestration
