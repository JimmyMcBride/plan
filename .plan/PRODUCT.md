# Product Direction

Date: 2026-04-15

## Thesis

`plan` is a local-first planning tool for AI-assisted software projects.

It exists to turn messy ideas into execution-ready plans that agents can follow reliably.

It does not own memory, retrieval, or context management. That belongs to a separate companion tool. `plan` owns planning only.

## Non-negotiables

- Local-first.
- All project planning material lives in `.plan/` at repo root.
- Strong skill integration.
- Release and GitHub workflow should follow a clean merge-to-release pattern.
- Default workflow should preserve:
  - brainstorm
  - epic
  - spec
  - stories

## Product principles

### 1. Default simple, advanced optional

New user path should feel obvious in minutes.

Power users should be able to add structure without switching tools.

### 2. Specs are the contract

Brainstorms are exploratory.
Epics define scope.
Specs become canonical.
Stories execute the approved spec.

### 3. Planning artifacts must be useful to agents

Every durable artifact should improve execution quality:

- clear scope
- explicit constraints
- verification expectations
- linked upstream source of truth

### 4. No planning theater

No story points.
No fake sprint rituals.
No stakeholder roleplay.
No PM dashboard cosplay.

### 5. Migrations only if they stay small

Use workspace upgrades only for:

- directory normalization
- template changes
- metadata repair
- new required fields

Avoid heavy schemas and avoid migrations as a product identity.

## Recommended core model

### Canonical hierarchy

1. Epic
2. Spec
3. Story

### Workflow entry

1. Brainstorm
2. Promote to epic
3. Write and approve spec
4. Split into stories

### Supporting surfaces

1. Roadmap
2. Dependency graph
3. Ready queue
4. External sync adapters

## Recommended `.plan/` file model

```text
.plan/
  PROJECT.md
  ROADMAP.md
  brainstorms/
  epics/
  specs/
  stories/
  .meta/
    workspace.json
    migrations.json
```

Notes:

- `PROJECT.md` holds enduring planning direction only.
- `ROADMAP.md` is the lightweight portfolio and ordering layer above epics.
- `brainstorms/` is divergent scratch space and promotion input, not a canonical planning level.
- `epics/` is scoped initiative layer.
- `specs/` is canonical execution contract layer.
- `stories/` is smallest tracked execution layer.
- `.meta/` is tool-owned state, not user-authored planning content.

## Recommended workflow

### Default path

1. `plan init`
2. `plan brainstorm start "idea"`
3. `plan epic promote <brainstorm>`
4. `plan spec edit <epic>`
5. `plan spec approve <epic>`
6. `plan story create <epic> "story title"`
7. `plan status`

Mental model:

- brainstorm = discovery
- epic = outcome boundary
- spec = canonical contract
- story = execution unit

### Key gates

- No stories before spec approval.
- Every story links back to one canonical spec.
- Stories should be execution-ready, not vague placeholders.
- Specs must include explicit non-goals and verification expectations.
- Closing an epic requires story completion or explicit abandoned/deferred decisions.

## Recommended CLI shape

### v1 core

- `plan init`
- `plan doctor`
- `plan update`
- `plan brainstorm ...`
- `plan epic ...`
- `plan spec ...`
- `plan story ...`
- `plan roadmap ...`
- `plan status`
- `plan skills install`

### Commands to delay

- external integrations
- hosted sync
- heavy dashboards
- database mode
- multi-user locks

## Recommended artifact expectations

### Brainstorm

Loose. Idea capture. Questions. raw notes. possible themes.

### Epic

Outcome, scope, why now, source links, current state.

### Spec

Problem, goals, non-goals, constraints, solution shape, flows, data/interfaces, risks, rollout, verification, story breakdown.

### Story

Concrete execution unit with:

- description
- acceptance criteria
- references
- verification steps
- status

## Migration stance

Recommendation: yes, but only as a small safety feature.

Rules:

- migrations must be idempotent
- migrations must be inspectable
- migrations must be reversible when possible
- migrations must only touch `plan`-managed surfaces
- `plan doctor` must explain current, pending, or broken state

This should feel like workspace repair, not ORM migration culture.

## What to copy from reference tools

### Copy from local-first reference tools

- local markdown ownership
- skill install model
- release workflow
- migration philosophy
- epic/spec/story gate

### Copy from `get-shit-done`

- verification-aware planning
- roadmap layer
- plan quality checks
- advanced workflows as opt-in

### Copy later from `beads`

- dependency graph
- ready queue
- branch-safe IDs if needed

## Initial roadmap

### v0.1

- initialize `.plan/`
- create/edit/list core artifacts
- lightweight `ROADMAP.md`
- promote brainstorm -> epic -> spec
- spec approval gate
- story creation and status
- skill installation
- doctor/update/migrate support

### v0.2

- roadmap management
- story verification fields
- plan quality checks
- story decomposition helpers

### v0.3

- dependencies between stories/epics
- ready view
- stronger local workflow ergonomics

### v0.4+

- external sync adapters
- GitHub/Jira/Linear export-sync
- small-team collaboration patterns

## Open questions for user

### 1. Roadmap in v1 or later?

Recommendation: yes in v1, but lightweight. One `ROADMAP.md`, not full phase machinery.

### 2. Story format: plain tracking or execution-ready?

Recommendation: execution-ready. Every story should include verification instructions, not just status.

### 3. IDs: human slugs only or stable IDs too?

Recommendation: human slugs in v1. Add stable IDs only when dependency graphs or merge collisions become real pain.

### 4. Should `plan` own approvals?

Recommendation: yes, but lightly. Status gates on specs and stories. No enterprise approval workflows.

### 5. Should `plan` import from other planning systems?

Recommendation: maybe later, but only after `plan` proves its own native model and only if imports can stay optional and explicit.
