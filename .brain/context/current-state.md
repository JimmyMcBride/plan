---
updated: "2026-04-27T15:44:41Z"
---
# Current State

<!-- brain:begin context-current-state -->
This file is a deterministic snapshot of the repository state at the last refresh.

## Repository

- Project: `plan`
- Root: `.`
- Runtime: `go`
- Go module: `plan`
- Current branch: `codex/codex-cloud-brain-setup`
- Default branch: `main`
- Remote: `https://github.com/JimmyMcBride/plan.git`
- Go test files: `561`

## Docs

- `README.md`
- `docs/gitflow.md`
- `docs/project-architecture.md`
- `docs/project-overview.md`
- `docs/project-workflows.md`
- `docs/using-plan.md`
<!-- brain:end context-current-state -->

## Local Notes

Add repo-specific notes here. `brain context refresh` preserves content outside managed blocks.

- On April 22, 2026, `./scripts/refresh-plan-develop-context.sh` reconciled checked-in GitHub planning metadata on `develop` and updated `.plan/.meta/github.json` so guide packet issues `#34`-`#36` are recorded as closed/merged with doc refs normalized to `develop`.
- On April 22, 2026, the GitHub collaboration foundation landed on `codex/github-collaboration-foundation`: `plan source show|set`, `plan discuss assess|promote`, GitHub Discussion assessment, draft-first promotion, initiative/spec issue orchestration, milestone creation, dependency wiring, and local planning mirrors in `.plan/.meta/github.json`.
- On April 23, 2026, PR `#55` review feedback tightened the GitHub collaboration foundation: promotion dependency edges now wire in a second pass after all issues exist, guide packets default blank `source_mode` to `local`, GraphQL responses now fail on `errors` and paginate Discussion comments, promotion drafts keep `proposed_spec_issues` as a stable empty array when not ready, and bullet parsing strips GitHub task-list markers before deriving spec titles.
- On April 23, 2026, workspace refresh stopped backfilling optional compatibility defaults into tracked metadata during `plan update`: `source_mode` and GitHub `planning` map now default in memory on read, so `./scripts/refresh-plan-develop-context.sh` no longer dirties `develop` just to normalize older `.plan/.meta/*.json` files.
- On April 25, 2026, GitHub promotion became fail-closed: explicit multi-spec sources that parse as fewer than two specs now return `needs_source_repair`, `plan discuss repair` owns canonical `## Specs` repair, promotion drafts include hard agent policy and fallback gating, 5+ spec apply/adopt requires `--project-decision`, `plan github adopt` recovers manual issue sets, and `plan check` detects Plan-labeled GitHub planning drift.
- On April 25, 2026, PR `#62` review comments tightened the fail-closed promotion branch: GitHub issue listing now raises on 1000-item truncation instead of silently missing Plan-labeled issues, `github adopt` validates output format before mutation, drift findings are deterministic and preserve milestone display casing, label ensuring only touches Plan-owned `plan:*` labels, and shell command quoting reuses a package-level regex.
- On April 27, 2026, Plan skill installation expanded from one bundled skill to two: `plan skills install` now installs both the base `plan` planning skill and the companion `plan-execute` execution rail skill, with manifests recording the installed skill name.
- On April 27, 2026, PR `#71` merged GitHub project workspace provisioning and item field setup into `develop`; review follow-up touched `internal/planning/collaboration.go`, `internal/planning/collaboration_test.go`, `internal/planning/github_client.go`, `internal/planning/github_client_test.go`, and `internal/planning/github_project_workspace.go` to require explicit project decisions for project references, accept `www.github.com` project URLs with trailing view paths, and cache issue node IDs from CLI issue payloads before falling back to GraphQL lookups.
