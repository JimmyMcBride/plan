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

Execution loop in GitHub mode:

1. Establish queue
2. Grab next ready issue or issues
3. Implement on a branch and open PR
4. Review and iterate until ready
5. Squash-merge
6. Return to the integration branch, pull latest, update and reconcile
7. Grab next ready issue and repeat

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

## GitHub Queue Workflow

For GitHub-backed stories, use this loop:

```bash
plan status --project .

# do issue work on a feature branch and merge the PR into the integration branch

./scripts/refresh-plan-develop-context.sh
```

Rules:

- use `plan status --project .` as the queue view
- take only issues shown in `ready_work`
- in this repo, normal work targets `develop`
- work on a feature branch, not on `develop`, `release/*`, or `main`
- squash-merge the PR when work is accepted unless release work needs a
  different merge strategy
- after merge into `develop`, run
  `./scripts/refresh-plan-develop-context.sh` before grabbing the next issue

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
