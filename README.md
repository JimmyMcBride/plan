# plan

`plan` is a local-first planning CLI for AI-assisted software work.

It keeps planning material in `.plan/` and focuses on one job: turning rough
ideas into shaped, execution-ready plans that agents can follow cleanly.

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
3. Promote to epic
4. Write and approve spec
5. Analyze the spec
6. Split into stories

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
```

User-authored planning material lives in:

- `.plan/PROJECT.md`
- `.plan/ROADMAP.md`
- `.plan/brainstorms/`
- `.plan/epics/`
- `.plan/specs/`
- `.plan/stories/`

Tool-owned state lives only in:

- `.plan/.meta/workspace.json`
- `.plan/.meta/migrations.json`

`plan update` may repair or normalize tool-owned state. It must not rewrite
user-authored planning notes just to migrate product direction.

## Quick Start

```bash
plan init --project .
plan brainstorm start --project . "Newsletter system"
plan brainstorm refine --project . newsletter-system
plan epic promote --project . newsletter-system
plan spec show --project . newsletter-system
plan spec analyze --project . newsletter-system
plan spec status --project . newsletter-system --set approved
plan story create --project . newsletter-system "Build template editor" \
  --criteria "Templates can be created and edited" \
  --verify "go test ./..."
plan status --project .
```

## Current Command Surface

- `plan init`
- `plan adopt`
- `plan doctor`
- `plan update`
- `plan brainstorm start|idea|show|refine`
- `plan epic create|promote|list|show`
- `plan spec show|edit|status|analyze`
- `plan story create|update|list|show`
- `plan roadmap show|edit`
- `plan check`
- `plan status`
- `plan skills install|targets`

## Roadmap Direction

Product phases, not release-tag numbers:

- `v4`: Planning Refinement Foundation
- `v5`: Planning Skills, Shaping, and Evals
- `v6`: Story Slicing and Execution Readiness
- `v7`: External Sync, only if the local loop clearly wins

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

## Release Flow

- Pull requests run `go test ./...` and `go build ./...` in CI.
- Every push to `main` tags the next patch release if `HEAD` is not already tagged.
- The release workflow builds platform archives and publishes a checksum file with the release assets.
- `scripts/install.sh` only falls back to a source build when no published release can be resolved. Download or checksum failures stay hard failures.

## Maintainers

- Keep pull request titles and descriptions release-note-friendly. The `## Release Notes` section in the PR template is the source of truth for published release changelogs.
- Include the verification commands you ran in the PR so the release notes have a clean audit trail.
- Use `scripts/next-release-tag.sh` if you need to preview the next patch tag locally.
