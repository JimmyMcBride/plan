## Summary
- describe the user-facing change in release-note language
- keep this focused on shipped behavior, not implementation mechanics

## Release Notes
- list the 1-5 highest-signal user-visible changes
- write these as human-readable bullets that can survive into GitHub-generated release notes
- if the PR is mainly a fix, say what was broken and what is now correct

## Verification
- go test ./...
- go build ./...

## Maintainer Notes
- if this PR changes `skills/plan/`, reinstall the local Plan skill before closeout:
  - `plan skills install --scope local --agent codex --project .`
