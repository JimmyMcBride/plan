# Project: plan

Created: 2026-04-16T05:33:06Z

## Vision

Build the best local-first planning tool for AI-assisted software projects.

`plan` should help indie developers and small teams turn rough ideas into
shaped, execution-ready specs and stories without PM theater, cloud lock-in, or
memory/context-management bloat.

## Principles

- Local-first and markdown-first.
- Planning only.
- Specs are the contract.
- Simple default flow, deeper shaping later.
- Improve the existing artifacts before inventing new ones.

## Constraints

- All durable planning material lives in `.plan/`.
- No hosted dependency is required for core workflows.
- No issue-tracker clone behavior in the planning core.
- External sync is later than local refinement quality.
- This repo uses Gitflow with `develop` as the active integration branch,
  `release/vX.Y.Z` as the release stabilization branch, and `main` as the
  release-only production branch.
- Protected branches are changed only through pull requests.

## Planning Rules

- Brainstorm is discovery, not a canonical hierarchy level.
- Specs are the canonical execution contract.
- Stories are created only after spec approval.
- Stories should be execution-ready and verification-aware.
- New passes should improve clarity, boundedness, and agent executability.
- If GitHub story mode is enabled, stories live in GitHub Issues rather than
  local markdown notes.
- GitHub execution follows a queue loop: establish ready work, grab the next
  issue or issues, ship a PR, review until ready, squash-merge into
  `develop`, refresh local `develop` from `origin/develop`, run `plan update`,
  reconcile, then grab the next ready work.

## Delivery Rules

- Normal ongoing work lands in `develop`.
- Official releases are cut from `develop` onto `release/vX.Y.Z`, then merged
  into `main`.
- Release fixes land in `develop` first, then are cherry-picked into the active
  `release/vX.Y.Z` branch.
- Production hotfixes must be merged back into `develop`.
- After each merge into `develop`, run
  `./scripts/refresh-plan-develop-context.sh` to refresh local `plan` context.

## Notes

- v4 focuses on refinement and simplification.
- v5 focuses on skill quality, shaping, and evals.
- v6 focuses on story slicing and critique.
- v7 is GitHub sync only if local planning quality clearly earns the added complexity.
- v8 focuses on guided co-planning and stage-by-stage execution handoffs.
