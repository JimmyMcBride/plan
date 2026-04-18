# Product Direction

Date: 2026-04-17

## Thesis

`plan` is a local-first shaping and execution-readiness engine for AI-assisted
software projects.

It should improve the quality of the existing flow more than it expands the
surface area of the product. The best next versions add better refinement,
analysis, shaping, slicing, and critique passes around `Brainstorm -> Epic ->
Spec -> Story`.

## Product Rules

- Local-first.
- Markdown-first.
- Planning only.
- All durable planning material lives in `.plan/`.
- Specs are the canonical execution contract.
- Stories are created only after spec approval.
- Stories should be execution-ready and verification-aware.
- Default workflow stays simple; advanced passes stay optional.

## Current Reset

The old power-first direction is no longer the product story.

That means:

- `ready` is not a headline feature
- dependency-heavy workflow features are not the next step
- external sync is explicitly later
- the next release work should favor refinement quality over coordination power

## Roadmap Phases

### v4: Planning Refinement Foundation

- converge the docs and templates on one thesis
- remove the older power-first command surface
- ship `plan brainstorm refine`
- ship `plan spec analyze`

### v5: Planning Skills, Shaping, and Evals

- add model-aware planning skill guidance
- add benchmark fixtures and rubric-based evals
- ship `plan brainstorm challenge`
- ship `plan epic shape`
- ship `plan spec checklist`

### v6: Story Slicing and Execution Readiness

- ship `plan story slice`
- ship `plan story critique`
- make the spec-to-story transition materially better

### v7: External Sync, Only If Still Needed

- GitHub first, only if the local loop is clearly winning
- `.plan` stays canonical
- external tools remain projections, not source of truth

## Artifact Expectations

### Brainstorm

Discovery material plus a durable refinement pass.

### Epic

Outcome boundary. Later shaping adds appetite, out-of-scope, and success signal.

### Spec

Canonical contract plus non-destructive analysis and checklist passes.

### Story

Small execution unit with acceptance criteria and verification. Later critique
and slicing make the boundary tighter, not larger.

## Release Position

Product phases are `v4+`. Release tags can remain `v0.x.y` until a separate
`1.0` decision is justified.
