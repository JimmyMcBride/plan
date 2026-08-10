---
updated: "2026-08-10T07:23:35Z"
---
# Project Architecture

<!-- brain:begin project-doc-architecture -->
Use this file for the structural shape of the repository.

## Internal Packages

- `internal/buildinfo/`
- `internal/notes/`
- `internal/planning/`
- `internal/skills/`
- `internal/templates/`
- `internal/workspace/`

## Architecture Notes

- Favor small package boundaries and explicit CLI/app wiring.
- Keep public CLI behavior stable; add internal seams only when they improve testability or safety.
- Treat generated project context as deterministic repo state, not LLM-authored prose.
- Treat session enforcement as the hard-control layer above soft context files.
<!-- brain:end project-doc-architecture -->

## Local Notes

- Schema-v3 local planning commands use `github.com/JimmyMcBride/brain` public
  Planning packages through `cmd/shared_planning.go`; the standalone command
  layer remains the host for flags, prompts, rendering, and legacy behavior.
- The bridge does not write Brain-specific event artifacts. Its compatibility
  event sink is intentionally process-local so the standalone `.plan/` disk
  contract stays unchanged.
- GitHub, hybrid, legacy epic/story, unsupported guide checkpoints, and invalid
  legacy spec repair paths remain standalone fallbacks.
- Brain must be pinned to a merged revision in `go.mod`; local `replace`
  directives are not allowed for the compatibility cutover.
