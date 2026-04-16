# plan

`plan` is a local-first planning CLI for AI-assisted software work.

It keeps planning material in `.plan/` inside the repo and focuses on one job:
turning rough ideas into execution-ready plans that agents can follow cleanly.

## Philosophy

- local-first
- markdown-first
- planning only
- simple default workflow
- deeper power available later

`plan` does not own memory, retrieval, or context management. That belongs to
other tools. `plan` owns planning.

## Core Model

Canonical hierarchy:

1. Epic
2. Spec
3. Story

Workflow entry:

1. Brainstorm
2. Promote to epic
3. Write and approve spec
4. Split into stories

Supporting surfaces:

- `PROJECT.md`
- `ROADMAP.md`
- optional future dependency and ready views

Simple default path:

1. brainstorm
2. promote
3. approve spec
4. create stories
5. execute

Advanced path stays optional. When a repo grows, you can add roadmap versions,
dependency blockers, ready-work views, Brain imports, and filtered status views
without changing the base model.

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
user-authored planning notes.

## Quick Start

```bash
plan init --project .
plan brainstorm start --project . "Newsletter system"
plan epic promote --project . newsletter-system
plan spec show --project . newsletter-system
plan spec status --project . newsletter-system --set approved
plan story create --project . newsletter-system "Build template editor"
plan status --project .
```

## Advanced Local Workflows

These stay optional. If you do not need them, ignore them.

- adopt an existing repo into a managed workspace:
  `plan adopt --project .`
- run structural planning checks:
  `plan check --project .`
- inspect version slices of the roadmap:
  `plan roadmap versions --project . --version v2`
- filter status for larger plans:
  `plan status --project . --version v3 --epic power-user-local-workflows --story-status todo`
- surface ready and blocked work:
  `plan ready --project .`
- narrow story lists by roadmap version:
  `plan story list --project . --version v3`
- inspect or import planning notes from a local Brain workspace:
  `plan import brain inspect --workspace ../brain`
  `plan import brain apply --project . --workspace ../brain --epic planning-and-brainstorming-ux`

The rule stays the same: use the advanced surfaces only when the simple default
stops being enough.

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

- Keep pull request titles and descriptions release-note-friendly. Generated GitHub release notes use merged PR metadata.
- Include the verification commands you ran in the PR so the release notes have a clean audit trail.
- Use `scripts/next-release-tag.sh` if you need to preview the next patch tag locally.
