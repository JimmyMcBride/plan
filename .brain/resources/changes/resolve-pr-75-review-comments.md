---
updated: "2026-05-04T00:00:00Z"
---
# Resolve PR 75 Review Comments

## Outcome

PR `#75` review comments were addressed on branch `codex/project-status-automation-reconcile`.

The follow-up hardened GitHub Project status automation by:

- returning non-nil `Values` maps for newly added Project items and defensively initializing maps in reconcile before assignment
- extending Project item lookup results with issue URL, title, and state so drift/reconcile can avoid a separate `GetIssue` roundtrip per tracked record
- allowing reconcile to keep repairing safe Project drift when an existing single-select field is missing newer unused options, while still skipping values that cannot be set safely
- preserving status-command missing-card validation by treating empty item ids as missing Project cards

## Verification

- `go test ./internal/planning ./cmd`
- `git diff --check`
- `go test ./...`
- `go build ./...`
- `go run . check --project .`
- `go test -race ./...`
