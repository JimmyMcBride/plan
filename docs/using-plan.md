# Using `plan`

This guide describes how to use `plan` as it exists right now.

It is based on the current command surface in the repo, not on older roadmap
ideas.

## Current Product State

Right now:

- the active planning model is spec-first
- brainstorms stay local and can bloom into idea docs or specs
- `initiative` is lightweight optional grouping metadata
- `plan spec execute` is the active execution entry point
- legacy `epic` and `story` commands still exist during the transition
- GitHub-backed issue execution remains available when you enable GitHub mode

The top of this guide reflects the active spec-first model. Some later sections
still document legacy compatibility commands while the migration is in flight.

## What `plan` Is

`plan` is a local-first planning CLI for software projects.

It stores planning material in `.plan/` and focuses on one job:

- turn rough ideas into shaped planning artifacts
- make specs stronger before implementation starts
- guide execution from approved specs without persisting tiny slice artifacts

`plan` does not handle memory, retrieval, or context management.

## Core Model

Active model:

1. `Brainstorm`
2. `Idea Doc` (optional)
3. `Spec`
4. runtime `Slice`

Optional grouping:

- `Initiative` for multi-spec outcomes
- GitHub milestone mapping when GitHub integration is enabled

Legacy `epic` and `story` objects remain available as compatibility surfaces,
but they are not the default active model anymore.

Workflow entry:

1. `Brainstorm`
2. `Refine`
3. `Challenge`
4. `Promote or shape into a spec`
5. `Write and approve spec`
6. `Analyze or checklist the spec`
7. `Assign initiative metadata when needed`
8. `Start spec execution`
9. `Work slices one commit at a time`

## Workspace Layout

`plan` keeps its durable planning material under `.plan/`:

```text
.plan/
  PROJECT.md
  ROADMAP.md
  brainstorms/
  ideas/
  archive/
  specs/
  .meta/
    workspace.json
    migrations.json
    github.json
```

Meaning:

- `PROJECT.md`: product direction and project rules
- `ROADMAP.md`: version/phase planning
- `brainstorms/`: discovery notes
- `ideas/`: optional durable idea docs
- `archive/`: preserved legacy epic/story-era planning material
- `specs/`: canonical execution contracts
- `.meta/`: tool-owned state only, including GitHub story metadata when enabled

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

Archive legacy epic/story hierarchy without touching active specs:

```bash
plan update --project . --archive-legacy
```

## Existing Repo Setup

If the repo already exists and you want `plan` to manage it:

```bash
plan adopt --project .
plan doctor --project .
```

## Optional: Enable GitHub Story Mode

If you want local planning but GitHub-backed issue execution during the
transition:

```bash
plan update --project .
plan github enable --project .
```

Preconditions:

- `gh` is installed
- `gh auth status` passes
- the repo has GitHub Issues enabled
- local story notes are not still active in `.plan/stories/`

When GitHub story mode is enabled:

- execution stories are created as GitHub Issues
- `.plan/.meta/github.json` becomes the local issue-state index
- initiative metadata can map execution work to GitHub milestones when titles
  match

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

### 10. Slice Stories From The Spec

If the approved spec already has a strong `Story Breakdown`, preview the first
pass slice set:

```bash
plan story slice --project . newsletter-system
```

This reads the canonical spec and produces a deterministic preview of the
candidate stories it can derive from `## Story Breakdown`.

Behavior:

- preview-first by default
- uses the story breakdown as the source of truth
- derives acceptance criteria and verification from the spec when needed
- protects against duplicate story creation on reruns

Apply the slice set:

```bash
plan story slice --project . newsletter-system --apply
```

This creates any missing story notes and rewrites `## Story Breakdown` with
linked checklist entries.

Manual `story create` still works. Slicing is an optional accelerator, not a
mandatory ceremony.

### 11. Create Stories Manually

Create a story from an approved spec:

```bash
plan story create --project . newsletter-system "Build template editor" \
  --body "Create the template editing flow." \
  --criteria "Templates can be created and edited" \
  --criteria "Template validation errors are visible" \
  --verify "go test ./..." \
  --verify "Manually verify the editor flow"
```

If GitHub story mode is enabled, the same command creates a GitHub Issue-backed
story instead of a local markdown story note.

Important rules:

- a story requires at least one acceptance criterion
- a story requires at least one verification step
- story creation is blocked until the spec is `approved`

### 11A. Spec Queue Workflow

Preferred execution loop is spec-first, not issue-first:

1. Establish queue with `plan status --project .`
2. Take the next approved spec from the queue
3. Run `plan story slice --project . <epic-slug>` to preview the slices
4. Run `plan story slice --project . <epic-slug> --apply` when the slice set is sound
5. Implement one slice
6. Review and verify that slice before committing it
7. Repeat slice-by-slice until the spec is done
8. Move to the next spec in queue if more queued specs remain
9. Open one PR when the queued specs for the branch are complete

Recommended commands:

```bash
plan status --project .

# take next approved spec
plan story slice --project . <epic-slug>
plan story slice --project . <epic-slug> --apply

# implement one slice
# run review + verification
# commit slice

# repeat until the spec is done
# then move to the next queued spec
```

Use this model:

- `plan status` = queue view
- approved spec = execution batch
- story slices = execution units inside the current spec
- review + verification happen before each slice commit
- PR = review for the completed queued spec batch

If GitHub story mode is enabled, reconcile after merge:

```bash
git switch <integration-branch>
git pull --ff-only origin <integration-branch>
plan update --project .
plan github reconcile --project . --update-visible
plan status --project .
```

If GitHub story mode is not enabled, use the same refresh without reconcile:

```bash
git switch <integration-branch>
git pull --ff-only origin <integration-branch>
plan update --project .
plan status --project .
```

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

### 12. Critique Story Readiness

Pressure-test a story before implementation:

```bash
plan story critique --project . build-template-editor
```

This writes the `## Critique` section in the story note:

- `Scope Fit`
- `Vertical Slice Check`
- `Hidden Prerequisites`
- `Verification Gaps`
- `Rewrite Recommendation`

Behavior:

- interactive, TTY-first
- additive to the story note
- returns non-zero when the recommendation is `rewrite` or `reslice`

### 13. Update Story Status

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

plan story slice --project . billing-export
plan story slice --project . billing-export --apply

plan status --project .
plan check --project .
```

## What `plan` Does Not Do Right Now

Current state means:

- no memory or context-management features
- no external tracker sync
- no cloud-first workflow

Those are roadmap questions, not current usage.

## `v6` Execution-Readiness Workflow

The main `v6` additions are about making the spec-to-story handoff stronger.

Recommended flow:

1. Approve the spec.
2. Make sure `## Story Breakdown` contains meaningful slice candidates.
3. Run `plan story slice --project . <epic-slug>` to preview the first pass.
4. Run `plan story slice --project . <epic-slug> --apply` when the preview is sound.
5. Run `plan story critique --project . <story-slug>` on the stories that need pressure-testing.
6. Run `plan check --project .` to validate the spec-to-story handoff.

`plan check` now looks for readiness problems such as:

- implementing specs with no story breakdown
- implementing specs with no child stories
- linked story breakdown entries that point to missing files
- story sets that exist but are not reflected in the canonical breakdown
- implementing specs whose stories are all still `todo`

## Practical Rules

- Start with the smallest useful pass.
- Use `refine` when the idea is fuzzy.
- Use `challenge` when the idea is too comfortable or too broad.
- Use `shape` when the epic boundary is weak.
- Use `analyze` for general spec pressure-testing.
- Use `checklist` when the spec has domain-specific risk.
- Approve the spec before creating or slicing stories.
- Use `story critique` when a slice feels too broad or verification-thin.
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
- `plan github`
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
plan github --help
```
