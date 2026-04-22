# Roadmap: plan

Created: 2026-04-17T00:00:00Z

## Overview

`plan` is resetting around planning refinement rather than older power-first
local workflows. The next few product phases focus on shaping quality, better
prompts and evals, stronger execution-readiness, and then explicit
source-of-truth flexibility beyond the default local backend.

## v4: Planning Refinement Foundation

Goal: Reset the product story, simplify the CLI, and ship the first two
high-leverage shaping passes.

- [x] Product Reset and Docs Convergence
- [x] CLI Surface Simplification
- [x] Guided Brainstorm Refinement
- [x] Spec Analysis Foundation

Summary:
- converge README, product docs, roadmap, project notes, templates, and skill text
- remove old power-first commands from the main CLI surface
- add `plan brainstorm refine`
- add `plan spec analyze`

## v5: Planning Skills, Shaping, and Evals

Goal: Make prompt quality, shaping passes, and benchmarked improvements
first-class product work.

- [x] Model-Aware Planning Skills
- [x] Benchmark Fixtures and Rubric Evals
- [x] Brainstorm Challenge
- [x] Epic Shaping
- [x] Spec Checklist Profiles

Summary:
- ship model-aware planning skill guidance
- benchmark prompt and workflow improvements before expanding the surface
- add the next shaping passes around brainstorms, epics, and specs

Current focus:
- move into `v6` story slicing and execution-readiness work
- keep the default path simple while the optional shaping passes deepen

## v6: Story Slicing and Execution Readiness

Goal: Make the spec-to-story transition materially better before reintroducing
larger-plan power.

- [x] Story Slice Preview and Apply Flow
- [x] Story Critique and Rejection Rules
- [x] Execution-Readiness Integration in `plan check`

Summary:
- add `plan story slice`
- add `plan story critique`
- keep the default path small while making optional story quality passes stronger

## v7: Backend Flexibility and GitHub Integration

Goal: Expand beyond local-only assumptions without losing the simple local
default.

- [ ] Source-of-Truth Backends and Ownership Model
- [ ] GitHub Planning Surfaces and Workflow Modeling
- [ ] GitHub-Backed Planning and Execution Loops

Summary:
- keep `local` as the default mode, not the only mode
- support explicit `local`, `github`, and `hybrid` source-of-truth backends
- make planning-layer ownership explicit instead of assuming `.plan/` always
  owns every durable artifact
- use GitHub as the first serious external planning backend

## v8: Guided Co-Planning System

Goal: Turn `plan` into a guided, stage-by-stage co-planner that walks users
from rough vision through story creation without losing artifact quality.

- [ ] Guided Session Engine and Resume
- [ ] Vision Intake and Brainstorm Co-Planning
- [ ] Guided Stage Handoffs and Artifact Writing
- [ ] Reopen Review and Roadmap Parking

Summary:
- keep guided conversation as the default planning mode
- ask users for vision and relevant docs up front instead of auto-scanning
- move stage-by-stage with recap plus `continue / refine / stop for now`
- persist chain-scoped sessions and mark downstream work `needs review` after upstream changes

## Ordering Notes

- fix the product story before adding more product surface
- benchmark quality improvements before stacking more passes
- keep the default path simple even as optional shaping modes deepen
- backend flexibility should preserve simple local defaults and explicit
  ownership boundaries

## Parking Lot

- generic ready queues
- dependency graphs as headline features
- Jira or Linear adapters
- hosted dashboards
- cloud-first collaboration
- memory, retrieval, or context-engineering features
