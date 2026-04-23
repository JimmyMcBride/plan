# plan

`plan` is a local-first-by-default, backend-flexible planning CLI for
AI-assisted software work.

It focuses on one job: turning rough ideas into shaped, execution-ready plans
that agents can follow cleanly. `.plan/` is the default local workspace, but
configured integrations can own persistent planning data in `github` or
`hybrid` modes.

## Philosophy

- local-first
- backend-flexible
- markdown-first
- planning only
- simple default workflow
- optional deeper shaping passes

`plan` does not own memory, retrieval, or context management. Pair it with a
companion tool such as [`brain`](https://github.com/JimmyMcBride/brain) if you
need that layer.

## Core Model

Active planning model:

1. Brainstorm session
2. Distilled issue body or idea doc (optional)
3. Spec
4. Execution slices at runtime

`initiative` is optional lightweight grouping for multiple specs. In GitHub
mode, an initiative can map to a milestone and an initiative issue.

Legacy `epic` and `story` commands still exist during the transition, but they
are no longer the active model the workspace reports by default.

Workflow entry:

1. Brainstorm locally or in GitHub Discussion
2. Refine
3. Challenge
4. Assess maturity and draft promotion
5. Promote or shape the work into a spec or initiative
6. Write and approve spec
7. Analyze or checklist the spec
8. Assign initiative metadata when needed
9. Start spec execution
10. Work the execution slices one commit at a time

Execution loop:

1. Establish spec queue
2. Take next approved spec
3. Start execution from the approved spec
4. Implement one slice
5. Review and verify that slice before committing it
6. Repeat until the spec is done
7. Move to the next spec in queue
8. Open one PR when the queued specs are complete

The default path stays small. New shaping passes should improve the same
artifacts rather than add new top-level planning objects.

## Source Of Truth Modes

`plan` now treats source-of-truth choice as an explicit part of the product
model.

- `local`: durable planning data lives in `.plan/`
- `github`: durable planning data can live in GitHub issues, projects, and
  milestones
- `hybrid`: ownership is split across `.plan/` and integrations

Rules:

- local remains the default
- brainstorm is a session, not a durable hierarchy layer
- collaborative brainstorming can start in GitHub Discussions
- persistent planning data may live locally or in integrations
- ownership must be explicit by planning layer
- today, local is the most complete backend and GitHub is the first external
  backend being actively shaped

## Default Local Workspace

```text
my-project/
  .plan/
    PROJECT.md
    ROADMAP.md
    brainstorms/
    ideas/
    archive/
    specs/
    .meta/
      workspace.json
      migrations.json
      github.json
```

When local owns those planning layers, user-authored material lives in:

- `.plan/PROJECT.md`
- `.plan/ROADMAP.md`
- `.plan/brainstorms/`
- `.plan/ideas/`
- `.plan/archive/` for preserved legacy material
- `.plan/specs/`

Tool-owned local integration state lives only in:

- `.plan/.meta/workspace.json`
- `.plan/.meta/migrations.json`
- `.plan/.meta/github.json` when GitHub integration is enabled

`plan update` may repair or normalize tool-owned state. Use
`plan update --archive-legacy` to move legacy `epics/` and `stories/` into
`.plan/archive/` without mutating the active spec-first surfaces.

In `github` or `hybrid` modes, persistent planning data may also live outside
the repo while `.plan/.meta/` keeps local integration state and migration
metadata.

## Quick Start

```bash
plan init --project .
plan source show --project .
plan brainstorm start --project . "Newsletter system"
plan brainstorm refine --project . newsletter-system
plan brainstorm challenge --project . newsletter-system
plan discuss assess --project . --brainstorm newsletter-system --format json
plan discuss promote --project . --brainstorm newsletter-system --format json
# local repo-backed promotion still uses the legacy compatibility path today
plan epic promote --project . newsletter-system
plan spec show --project . newsletter-system
plan spec analyze --project . newsletter-system
plan spec checklist --project . newsletter-system --profile general
plan spec initiative --project . newsletter-system --set guide-packet-foundation --title "Guide Packet Foundation"
plan spec status --project . newsletter-system --set approved
plan spec execute --project . newsletter-system
plan status --project .
plan check --project .
```

GitHub collaborative path:

```bash
plan source set --project . github
plan discuss assess --project . --discussion 49 --format json
plan discuss promote --project . --discussion 49 --format json
plan discuss promote --project . --discussion 49 --apply --confirm --target github --format json
```

Full guide:

- [Using plan](docs/using-plan.md)

## Current Command Surface

- `plan init`
- `plan adopt`
- `plan doctor`
- `plan update`
- `plan source show|set`
- `plan brainstorm start|idea|show|refine`
- `plan brainstorm challenge`
- `plan discuss assess|promote`
- `plan guide current|show`
- `plan epic create|promote|list|show|shape` for legacy compatibility during migration
- `plan spec show|edit|status|analyze|checklist|initiative|execute|handoff`
- `plan story create|update|list|show|slice|critique` for legacy compatibility during migration
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
plan spec execute --project . <spec-slug>

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
- start execution from one approved spec before coding
- complete one slice at a time
- review and verify each slice before committing that slice
- once the current spec is done, move to the next queued spec
- in this repo, normal work targets `develop`
- work on a feature branch, not on `develop`, `release/*`, or `main`
- open one PR after the queued specs for that branch are complete
- if GitHub integration is enabled, run
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

`plan` remains the planning control plane. Depending on the configured backend,
durable planning truth may live in `.plan/`, GitHub, or both. Brain is only
for context, retrieval, and session hygiene when present.

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
