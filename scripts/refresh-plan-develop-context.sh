#!/bin/sh

set -eu

PROJECT_DIR="${1:-.}"

die() {
  printf 'refresh-plan-develop-context: %s\n' "$1" >&2
  exit 1
}

need_clean_worktree() {
  if [ -n "$(git status --porcelain)" ]; then
    die "clean worktree required before switching to develop"
  fi
}

ensure_local_develop() {
  if git show-ref --verify --quiet refs/heads/develop; then
    git switch develop >/dev/null
    git branch --set-upstream-to=origin/develop develop >/dev/null 2>&1 || true
    return
  fi
  git switch -c develop --track origin/develop >/dev/null
}

github_backend_enabled() {
  workspace_file="${PROJECT_DIR}/.plan/.meta/workspace.json"
  [ -f "${workspace_file}" ] || return 1
  grep -q '"story_backend"[[:space:]]*:[[:space:]]*"github"' "${workspace_file}"
}

need_clean_worktree
git fetch origin
ensure_local_develop
git pull --ff-only origin develop

go run . update --project "${PROJECT_DIR}"

if github_backend_enabled; then
  go run . github reconcile --project "${PROJECT_DIR}" --update-visible
fi

go run . status --project "${PROJECT_DIR}"
