# Using `plan`

This guide describes how to use `plan` as it exists right now.

It is based on the current command surface on `main`, not on older roadmap
ideas.

## Current Product State

Right now:

- the shipped CLI surface is `v5`
- `v6` is the next approved backlog
- this branch carries the `v6` planning artifacts, but not the `v6` feature
  implementation yet

So this guide covers what you can do today, and the end of the guide calls out
what is still missing for `v6`.

## What `plan` Is

`plan` is a local-first planning CLI for software projects.

It stores planning material in `.plan/` and focuses on one job:

- turn rough ideas into shaped planning artifacts
- make specs stronger before implementation starts
- make stories execution-ready before coding begins

`plan` does not handle memory, retrieval, or context management.

## Core Model

Canonical hierarchy:

1. `Epic`
2. `Spec`
3. `Story`

Workflow entry:

1. `Brainstorm`
2. `Refine`
3. `Challenge`
4. `Promote to epic`
5. `Shape the epic`
6. `Write and approve spec`
7. `Analyze or checklist the spec`
8. `Create and execute stories`

## Workspace Layout

`plan` keeps its durable planning material under `.plan/`:

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

Meaning:

- `PROJECT.md`: product direction and project rules
- `ROADMAP.md`: version/phase planning
- `brainstorms/`: discovery notes
- `epics/`: outcome boundaries
- `specs/`: canonical execution contracts
- `stories/`: execution-ready slices
- `.meta/`: tool-owned state only

## New Repo Setup

Initialize a new workspace:

```bash
plan init --project .
```

Check workspace health:

```bash
plan doctor --project .
```

Repair or normalize tool-owned state:

```bash
plan update --project .
```

## Existing Repo Setup

If the repo already exists and you want `plan` to manage it:

```bash
plan adopt --project .
plan doctor --project .
```

## Step-By-Step Workflow

### 1. Start a Brainstorm

Create a brainstorm:

```bash
plan brainstorm start --project . "Newsletter system"
```

Add notes or ideas:

```bash
plan brainstorm idea --project . newsletter-system --body "Use versioned templates"
plan brainstorm idea --project . newsletter-system --section constraints --body "Keep it local-first"
plan brainstorm idea --project . newsletter-system --section open-questions --body "How should previews work?"
```

Show the brainstorm:

```bash
plan brainstorm show --project . newsletter-system
```

### 2. Refine the Brainstorm

Run the guided refinement pass:

```bash
plan brainstorm refine --project . newsletter-system
```

This writes the `## Refinement` section in the brainstorm note:

- `Problem`
- `User / Value`
- `Appetite`
- `Remaining Open Questions`
- `Candidate Approaches`
- `Decision Snapshot`

Behavior:

- interactive, TTY-first
- saves after each cluster
- reruns are resumable

### 3. Challenge the Brainstorm

Pressure-test the idea before it hardens:

```bash
plan brainstorm challenge --project . newsletter-system
```

This writes the `## Challenge` section:

- `Rabbit Holes`
- `No-Gos`
- `Assumptions`
- `Likely Overengineering`
- `Simpler Alternative`

Use this when you want to force scope discipline before promotion.

### 4. Promote to an Epic

If the brainstorm is ready, promote it:

```bash
plan epic promote --project . newsletter-system
```

This creates:

- `.plan/epics/newsletter-system.md`
- `.plan/specs/newsletter-system.md`

You can also create an epic directly without a brainstorm:

```bash
plan epic create --project . "Newsletter system"
```

List epics:

```bash
plan epic list --project .
```

Show an epic:

```bash
plan epic show --project . newsletter-system
```

### 5. Shape the Epic

Run the epic shaping pass:

```bash
plan epic shape --project . newsletter-system
```

This writes the `## Shape` section:

- `Appetite`
- `Outcome`
- `Scope Boundary`
- `Out of Scope`
- `Success Signal`

It also mirrors key shape output back into the epic summary sections where it
helps readability.

### 6. Work the Spec

Show the canonical spec:

```bash
plan spec show --project . newsletter-system
```

Edit the spec with your editor:

```bash
plan spec edit --project . newsletter-system
```

Or replace the body directly:

```bash
plan spec edit --project . newsletter-system --body "$(cat my-spec.md)"
```

The spec is the canonical execution contract. It should contain:

- `Why`
- `Problem`
- `Goals`
- `Non-Goals`
- `Constraints`
- `Solution Shape`
- `Flows`
- `Data / Interfaces`
- `Risks / Open Questions`
- `Rollout`
- `Verification`
- `Story Breakdown`

### 7. Analyze the Spec

Run the general analysis pass:

```bash
plan spec analyze --project . newsletter-system
```

This writes the `## Analysis` section with findings for:

- `Missing Constraints`
- `Success Criteria Gaps`
- `Hidden Dependencies`
- `Risk Gaps`
- `What/Why vs How Leakage`
- `Recommended Revisions`

Behavior:

- non-destructive to the canonical spec sections
- returns non-zero when blocking findings exist

### 8. Run a Spec Checklist

Run a profile-specific pass:

```bash
plan spec checklist --project . newsletter-system --profile general
```

Current profiles:

- `general`
- `ui-flow`
- `api-integration`
- `data-migration`

This writes the `## Checklist` section in the spec and stores results under the
chosen profile heading.

Behavior:

- non-destructive to canonical sections
- deterministic for the same input
- returns non-zero when the selected profile reports blocking issues

### 9. Approve the Spec

Stories can only be created from an approved spec.

Approve it:

```bash
plan spec status --project . newsletter-system --set approved
```

Current spec statuses:

- `draft`
- `approved`
- `implementing`
- `done`

### 10. Create Stories

Create a story from an approved spec:

```bash
plan story create --project . newsletter-system "Build template editor" \
  --body "Create the template editing flow." \
  --criteria "Templates can be created and edited" \
  --criteria "Template validation errors are visible" \
  --verify "go test ./..." \
  --verify "Manually verify the editor flow"
```

Important rules:

- a story requires at least one acceptance criterion
- a story requires at least one verification step
- story creation is blocked until the spec is `approved`

Show a story:

```bash
plan story show --project . build-template-editor
```

List stories:

```bash
plan story list --project .
plan story list --project . --epic newsletter-system
plan story list --project . --status blocked
```

### 11. Update Story Status

Update story progress:

```bash
plan story update --project . build-template-editor --status in_progress
plan story update --project . build-template-editor --status done
```

Append more detail later:

```bash
plan story update --project . build-template-editor \
  --criteria "Editor preserves template formatting" \
  --verify "Run template editor tests"
```

Current story statuses:

- `todo`
- `in_progress`
- `blocked`
- `done`

## Quality Commands

Run checks across the repo:

```bash
plan check --project .
```

Or narrow checks to one artifact:

```bash
plan check epic newsletter-system --project .
plan check spec newsletter-system --project .
plan check story build-template-editor --project .
```

Use status to see overall project planning progress:

```bash
plan status --project .
```

## Roadmap Commands

Show roadmap:

```bash
plan roadmap show --project .
```

Edit roadmap:

```bash
plan roadmap edit --project .
```

## Skill Installation

Install the `plan` skill globally:

```bash
plan skills install --scope global --agent codex
```

Install locally in the repo:

```bash
plan skills install --scope local --agent codex --project .
```

Preview targets without writing:

```bash
plan skills targets --scope both --agent codex --project .
```

You can repeat `--agent` for multiple targets.

## End-To-End Example

```bash
plan init --project .

plan brainstorm start --project . "Billing export"
plan brainstorm idea --project . billing-export --body "Export billing data to an external API"
plan brainstorm refine --project . billing-export
plan brainstorm challenge --project . billing-export

plan epic promote --project . billing-export
plan epic shape --project . billing-export

plan spec show --project . billing-export
plan spec analyze --project . billing-export
plan spec checklist --project . billing-export --profile api-integration
plan spec status --project . billing-export --set approved

plan story create --project . billing-export "Trigger export job" \
  --criteria "Export job can be triggered from billing UI" \
  --verify "Run focused billing export tests"

plan story create --project . billing-export "Deliver export payload" \
  --criteria "Payload matches the external API contract" \
  --verify "Validate payload against fixture contract"

plan status --project .
plan check --project .
```

## What `plan` Does Not Do Right Now

Current state means:

- no memory or context-management features
- no external tracker sync
- no story slicing command yet
- no story critique command yet
- no cloud-first workflow

Those are roadmap questions, not current usage.

## What Is Left For `v6`

`v6` is not feature-complete yet. The approved backlog on this branch is:

### Story Slice Preview and Apply Flow

- add the story slice candidate model and preview formatting
- implement `plan story slice`
- add rerun and duplicate-protection coverage

### Story Critique and Rejection Rules

- add the `## Critique` section schema to story notes
- implement `plan story critique`
- add critique docs and test coverage

### Execution-Readiness Integration in `plan check`

- add spec-to-story readiness rules in `plan check`
- add cross-artifact readiness coverage
- document the `v6` execution-readiness workflow

In practical terms, `v6` becomes feature-complete when these three capabilities
exist together:

- a spec can be sliced into candidate stories with a preview-first flow
- a story can be critiqued and given keep/rewrite/reslice guidance
- `plan check` can validate the handoff from spec to stories across artifacts

## Practical Rules

- Start with the smallest useful pass.
- Use `refine` when the idea is fuzzy.
- Use `challenge` when the idea is too comfortable or too broad.
- Use `shape` when the epic boundary is weak.
- Use `analyze` for general spec pressure-testing.
- Use `checklist` when the spec has domain-specific risk.
- Approve the spec before creating stories.
- Keep stories small, concrete, and verification-aware.

## Current Command Surface

Top-level commands available today:

- `plan init`
- `plan adopt`
- `plan doctor`
- `plan update`
- `plan brainstorm`
- `plan epic`
- `plan spec`
- `plan story`
- `plan roadmap`
- `plan check`
- `plan status`
- `plan skills`

If you need the exact current help text, run:

```bash
plan --help
plan brainstorm --help
plan epic --help
plan spec --help
plan story --help
```
