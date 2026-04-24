# `plan` Gitflow

This document is the gitflow source of truth for the `plan` repository.

## Branch Roles

- `develop` is the active long-lived integration branch and the default target
  for routine work.
- `release/vX.Y.Z` is a protected release stabilization branch cut from
  `develop` when an official release is being prepared.
- `main` is the protected production branch. Merges into `main` remain the
  release event that triggers automatic publishing.

## Protected Branches

- `develop`, `release/*`, and `main` are protected branches.
- Never push directly to `develop`, any `release/*` branch, or `main`.
- Never delete `develop`, any `release/*` branch, or `main`.
- All changes to protected branches land through pull requests.

Current repo configuration:

- `develop` exists as the active integration branch.
- GitHub rules protect `develop` and `release/*` with pull-request-only updates,
  required CI checks, no force pushes, and no deletions.
- `main` remains separately protected and keeps the existing release publish
  behavior unchanged.

## Normal Development Flow

1. Start feature or bug-fix work from the current integration branch, usually
   `develop`.
2. Open routine pull requests into `develop`.
3. Do not merge routine work directly into `main`.
4. When a release is ready, cut `release/vX.Y.Z` from the current `develop`.
5. Stabilize and validate the release branch as needed.
6. Open a pull request from `release/vX.Y.Z` into `main`.
7. Merge into `main` only when ready to publish the official version.

## Release Stabilization

1. If a fix is needed while a release branch is open, land the fix in
   `develop` first.
2. After the fix exists in `develop`, cherry-pick the exact commit into the
   relevant `release/vX.Y.Z` branch.
3. Do not make one-off fixes directly on the release branch without first
   putting the equivalent fix into `develop`.
4. Treat `develop` as the long-term source of truth for future work.

## Hotfixes

1. If an urgent production-only fix is required, branch from the active
   `release/vX.Y.Z` branch or from `main`, whichever best reflects production.
2. Use the same pull-request-only process for that hotfix path.
3. After the production fix is prepared or merged, make sure the equivalent fix
   is merged back into `develop`.
4. `develop` must end up containing every production fix.

## Release Branch Retention

- Preserve release branches as historical snapshots of what was prepared for
  each official version.
- Keep them available for inspection, regression history, and later debugging
  questions.

## `plan` Maintenance Loop

After any pull request is merged into `develop`:

1. Fetch latest remote state.
2. Check out the updated `develop` branch from `origin/develop`.
3. Refresh `plan` project context from the latest `develop` state.

Repo helper:

```bash
./scripts/refresh-plan-develop-context.sh
```

That helper:

- fetches latest remote refs
- checks out and fast-forwards local `develop`
- runs `go run . update --project .`
- runs `go run . github reconcile --project . --update-visible` when GitHub
  story mode is enabled
- prints the latest `plan status --project .`

## Operational Defaults

- Treat `develop` as the active branch between releases.
- Treat `main` as release-only.
- Treat `release/vX.Y.Z` as the controlled bridge from `develop` to `main`.
- Use semantic version naming for release branches, exactly
  `release/vX.Y.Z`.
- If a choice is ambiguous, prefer the path that keeps `develop` current and
  keeps direct changes off protected branches.
