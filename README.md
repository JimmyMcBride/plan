# plan

`plan` is a local-first planning CLI for AI-assisted software work.

It keeps planning material in `.plan/` and focuses on one job: turning rough
ideas into shaped, execution-ready plans that agents can follow cleanly.

If GitHub story mode is enabled, brainstorms, epics, and specs stay local in
`.plan/`, while stories execute as GitHub Issues.

## Philosophy

- local-first
- markdown-first
- planning only
- simple default workflow
- optional deeper shaping passes

`plan` does not own memory, retrieval, or context management. Pair it with a
companion tool such as [`brain`](https://github.com/JimmyMcBride/brain) if you
need that layer.

## Core Model

Canonical hierarchy:

1. Epic
2. Spec
3. Story

Workflow entry:

1. Brainstorm
2. Refine
3. Challenge
4. Promote to epic
5. Shape the epic
6. Write and approve spec
7. Analyze or checklist the spec
8. Slice into stories
9. Critique story readiness

Execution loop:

1. Establish spec queue
2. Take next approved spec
3. Slice the spec into execution-ready stories
4. Implement one slice
5. Review and verify that slice before committing it
6. Repeat until the spec is done
7. Move to the next spec in queue
8. Open one PR when the queued specs are complete

The default path stays small. New shaping passes should improve the same
artifacts rather than add new top-level planning objects.

## Workspace

```text
my-project/
  .plan/
    PROJECT.md
    ROADMAP.md
    brainstorms/
    epics/
    specs/
    stories/
    .meta/
      workspace.json
      migrations.json
      github.json
```

User-authored planning material lives in:

- `.plan/PROJECT.md`
- `.plan/ROADMAP.md`
- `.plan/brainstorms/`
- `.plan/epics/`
- `.plan/specs/`
- `.plan/stories/` in local story mode

Tool-owned state lives only in:

- `.plan/.meta/workspace.json`
- `.plan/.meta/migrations.json`
- `.plan/.meta/github.json` when GitHub story mode is enabled

`plan update` may repair or normalize tool-owned state. It must not rewrite
user-authored planning notes just to migrate product direction.

## Quick Start

```bash
plan init --project .
plan brainstorm start --project . "Newsletter system"
plan brainstorm refine --project . newsletter-system
plan brainstorm challenge --project . newsletter-system
plan epic promote --project . newsletter-system
plan epic shape --project . newsletter-system
plan spec show --project . newsletter-system
plan spec analyze --project . newsletter-system
plan spec checklist --project . newsletter-system --profile general
plan spec status --project . newsletter-system --set approved
plan story slice --project . newsletter-system
plan story slice --project . newsletter-system --apply
plan story critique --project . build-template-editor
plan status --project .
plan check --project .
```

Full guide:

- [Using plan](docs/using-plan.md)

## Current Command Surface

- `plan init`
- `plan adopt`
- `plan doctor`
- `plan update`
- `plan brainstorm start|idea|show|refine`
- `plan brainstorm challenge`
- `plan epic create|promote|list|show|shape`
- `plan spec show|edit|status|analyze|checklist`
- `plan story create|update|list|show|slice|critique`
- `plan github enable|reconcile`
- `plan roadmap show|edit`
- `plan check`
- `plan status`
- `plan skills install|targets`

## Spec Queue Workflow

Use this loop when implementing planned work:

```bash
plan status --project .

# take next approved spec
plan story slice --project . <epic-slug>
plan story slice --project . <epic-slug> --apply

# implement one slice
# review + verify slice
# commit slice

# repeat until spec done
# move to next queued spec

# once queued specs are done, open one PR
```

Rules:

- use `plan status --project .` as the queue view
- queue work at the spec level, not the single-issue level
- slice one approved spec into execution-ready stories before coding
- complete one slice at a time
- review and verify each slice before committing that slice
- once the current spec is done, move to the next queued spec
- in this repo, normal work targets `develop`
- work on a feature branch, not on `develop`, `release/*`, or `main`
- open one PR after the queued specs for that branch are complete
- if GitHub story mode is enabled, run
  `./scripts/refresh-plan-develop-context.sh` after merge before taking more
  queue work

## Repo Gitflow

This repo uses Gitflow with `develop` as the active integration branch,
`release/vX.Y.Z` as the protected stabilization branch, and `main` as the
release-only production branch.

Normal work flows into `develop`. Official releases are cut from `develop` onto
`release/vX.Y.Z`, then merged into `main` to publish.

Release and maintenance rules live in [docs/gitflow.md](docs/gitflow.md).

## Roadmap Direction

Product phases, not release-tag numbers:

- `v4`: Planning Refinement Foundation
- `v5`: Planning Skills, Shaping, and Evals
- `v6`: Story Slicing and Execution Readiness
- `v7`: External Sync, only if the local loop clearly wins
- `v8`: Guided Co-Planning System

Release tags can stay in `v0.x.y` semver until a separate `1.0` decision.

## Install

Unix shell:

```bash
curl -fsSL https://raw.githubusercontent.com/JimmyMcBride/plan/main/scripts/install.sh | sh
```

Build from source:

```bash
git clone https://github.com/JimmyMcBride/plan.git
cd plan
go build -o plan .
install -Dm0755 plan ~/.local/bin/plan
```

## Install The Plan Skill

Global:

```bash
plan skills install --scope global --agent codex
```

Project-local:

```bash
plan skills install --scope local --agent codex --project .
```

Preview install targets:

```bash
plan skills targets --scope both --agent codex --project .
```

## Codex Cloud + Brain

If you want a Codex cloud environment for this repo to have optional access to
[`brain`](https://github.com/JimmyMcBride/brain), point the environment setup
step at:

```bash
./scripts/setup-codex-cloud.sh
```

That script:

- installs a repo-local Brain binary at `.codex/bin/brain`
- installs the repo-local Brain skill for Codex at `.codex/skills/brain`
- leaves Brain optional when the repo does not contain a `.brain/` workspace

`plan` remains the planning source of truth. Brain is only for context,
retrieval, and session hygiene when present.

## Evaluating Prompt And Workflow Changes

`v5` adds a local benchmark and rubric harness for maintainers. `v6` adds
story slicing, critique, and stronger spec-to-story readiness checks. The
benchmark workflow is test-driven:

```bash
go test ./internal/planning -run TestBenchmarkFixturesSatisfyMinimumScores
go test ./internal/planning -run TestRubricEvaluationIsDeterministic
```

The fixtures live under `testdata/evals/fixtures/` and the rubric code lives in
`internal/planning/evals.go`.

## Release Flow

- Pull requests run `go test ./...` and `go build ./...` in CI.
- `develop` is the default pull-request target for routine work.
- Official releases cut `release/vX.Y.Z` from `develop`, stabilize there, then
  merge into `main`.
- Every push to `main` still tags the next patch release if `HEAD` is not
  already tagged.
- The release workflow builds platform archives and publishes a checksum file with the release assets.
- `scripts/install.sh` only falls back to a source build when no published release can be resolved. Download or checksum failures stay hard failures.

## Maintainers

- Use [docs/gitflow.md](docs/gitflow.md) as the branching source of truth for
  this repo.
- Keep pull request titles and descriptions release-note-friendly. The `## Release Notes` section in the PR template is the source of truth for published release changelogs.
- Include the verification commands you ran in the PR so the release notes have a clean audit trail.
- Use `scripts/next-release-tag.sh` if you need to preview the next patch tag locally.
