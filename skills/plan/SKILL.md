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
- `plan epic create --project . "<title>"`
- `plan epic promote --project . <brainstorm-slug>`
- `plan spec show --project . <epic-slug>`
- `plan spec analyze --project . <epic-slug>`
- `plan spec status --project . <epic-slug> --set approved`
- `plan story create --project . <epic-slug> "<title>" --criteria "<criterion>" --verify "<step>"`
- `plan story update --project . <story-slug> --status in_progress`
- `plan roadmap show --project .`
- `plan status --project .`

## Rules

- Brainstorms are discovery material, not the canonical hierarchy.
- Canonical hierarchy is `Epic -> Spec -> Story`.
- `brainstorm refine` should reduce ambiguity before promotion.
- `spec analyze` should pressure-test a spec without rewriting its canonical sections.
- Keep roadmap guidance lightweight.
- Do not add tasks beneath stories as first-class objects unless the project explicitly asks for that system.
- Keep planning separate from memory, retrieval, or context management systems.
