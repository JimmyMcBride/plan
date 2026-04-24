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

Add repo-specific notes here. `brain context refresh` preserves content outside managed blocks.
