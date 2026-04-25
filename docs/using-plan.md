# Using `plan`

This guide describes how to use `plan` as it exists right now.

It is based on the current command surface in the repo, not on older roadmap
ideas.

## Current Product State

Right now:

- the active planning model is spec-first
- source-of-truth backends can be `local`, `github`, or `hybrid`
- brainstorms can start locally or in GitHub Discussions
- `initiative` is lightweight optional grouping metadata
- `plan spec execute` is the active execution entry point
- `plan guide current|show` emits live brainstorm and collaboration guide packets for agent runtimes
- `plan discuss assess|promote` ships the GitHub collaboration foundation
- `plan source show|set` makes backend ownership explicit
- legacy `epic` and `story` commands still exist during the transition
- GitHub integration is the first external backend being actively shaped

The top of this guide reflects the active spec-first model. Some later sections
still document legacy compatibility commands while the migration is in flight.

## What `plan` Is

`plan` is a local-first-by-default, backend-flexible planning CLI for software
projects.

It focuses on one job:

- turn rough ideas into shaped planning artifacts
- make specs stronger before implementation starts
- guide execution from approved specs without persisting tiny slice artifacts

`.plan/` is the default local workspace, but configured integrations can own
persistent planning data in `github` or `hybrid` modes.

`plan` does not handle memory, retrieval, or context management.

## Core Model

Active model:

1. `Brainstorm`
2. distilled issue body or `Idea Doc` (optional)
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
4. `Assess maturity and draft promotion`
5. `Promote or shape into a spec`
6. `Write and approve spec`
7. `Analyze or checklist the spec`
8. `Assign initiative metadata when needed`
9. `Start spec execution`
10. `Work slices one commit at a time`

## Source Of Truth Modes

`plan` now supports three source-of-truth modes:

- `local`: durable planning data lives in `.plan/`
- `github`: durable planning data can live in GitHub issues, projects, and
  milestones
- `hybrid`: ownership is split across `.plan/` and integrations

Rules:

- local is still the default
- brainstorm is a session, not a durable hierarchy layer
- collaborative shaping can also start in GitHub Discussions
- persistent planning data may live locally or in integrations
- ownership must be explicit by planning layer
- today, local is the most complete shipped backend and GitHub is the first
  external backend being actively shaped

## Default Local Workspace Layout

When local owns those planning layers, `plan` keeps durable material under
`.plan/`:

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

In `github` or `hybrid` modes, persistent planning data may also live outside
the repo while `.plan/.meta/` keeps local integration state and migration
metadata.

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

## Optional: Enable GitHub Integration

If you want to shape work with GitHub as part of the source-of-truth model:

```bash
plan update --project .
plan github enable --project .
```

Preconditions:

- `gh` is installed
- `gh auth status` passes
- the repo has GitHub Issues enabled
- GitHub Discussions should be enabled if you want `GitHub collaborative mode`

Current shipped GitHub support includes:

- GitHub Discussion assessment
- promotion drafting from a brainstorm or Discussion
- initiative/spec issue creation for `github` and `hybrid` targets
- milestone creation for multi-spec promotion
- parent/sub-issue and `blocked by` relationship wiring
- local mirror metadata in `.plan/.meta/github.json`

Full end-to-end GitHub-native spec execution is not finished yet. Repo-backed
`plan spec ...` execution remains the strongest shipped execution path.

When the current GitHub backend is enabled:

- promoted initiative and spec issues can become canonical planning artifacts
- GitHub Discussions can act as the collaborative brainstorm surface
- `.plan/.meta/github.json` becomes the local integration-state index
- initiative metadata can map multi-spec work to GitHub milestones

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

If you are using an external agent during a guided brainstorm, ask `plan` for
the live stage contract:

```bash
plan guide current --project . --format json
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

### 4. Pick An Entry Mode

`plan` now supports two collaboration entry modes:

- `local promotion mode`: brainstorm locally, then assess and promote later
- `GitHub collaborative mode`: shape the idea in a GitHub Discussion, then
  promote that Discussion into issue-backed planning artifacts

Inspect the current backend mode:

```bash
plan source show --project .
```

Switch to GitHub ownership when you want GitHub to own the promoted planning
artifacts:

```bash
plan source set --project . github
```

Current modes:

- `local`
- `github`
- `hybrid`

### 5. Assess Maturity Before Promotion

Assess a local brainstorm:

```bash
plan discuss assess --project . --brainstorm newsletter-system --format json
```

Assess a GitHub Discussion:

```bash
plan discuss assess --project . --discussion 49 --format json
```

The assessment decides whether the source is:

- `not_ready`
- `needs_source_repair`
- `ready_single_spec`
- `ready_multi_spec`

The JSON output includes:

- source mode and entry mode
- maturity strengths and gaps
- recommended path
- suggested issue titles
- an initial dependency guess
- a blocking repair command when explicit multi-spec intent cannot be parsed

If the source asks for multiple spec issues but Plan cannot parse at least two
spec titles, repair the source instead of creating GitHub issues manually:

```bash
plan discuss repair --project . --brainstorm newsletter-system --spec "Template CRUD" --spec "Preview API" --format json
plan discuss repair --project . --discussion 49 --spec "Template CRUD" --spec "Preview API" --confirm --format json
```

### 6. Review The Promotion Draft

Preview the promotion plan for a local brainstorm:

```bash
plan discuss promote --project . --brainstorm newsletter-system --format json
```

Preview the promotion plan for a GitHub Discussion:

```bash
plan discuss promote --project . --discussion 49 --format json
```

The draft tells you:

- whether the work should stay `single_spec` or fan out to `multi_spec`
- the proposed initiative issue, if needed
- the proposed spec issues
- parent/sub-issue grouping
- `blocked by` dependency suggestions
- whether a milestone should be created
- whether a project should be recommended
- whether any spec should start with `needs-refinement`
- the agent policy that forbids manual GitHub planning mutations unless Plan
  emits `manual_fallback_allowed=true`

Rules:

- single-spec promotion creates one spec issue and no initiative issue
- multi-spec promotion creates an initiative issue plus spec issues
- multi-spec promotion always creates a milestone
- the project prompt appears at `5+` specs or earlier when coordination is
  clearly messy

### 7. Apply Promotion To GitHub Or Hybrid Ownership

Once the draft looks right, apply it:

```bash
plan discuss promote --project . --discussion 49 --apply --confirm --target github --format json
```

You can also promote a local brainstorm into GitHub or hybrid ownership:

```bash
plan discuss promote --project . --brainstorm newsletter-system --apply --confirm --target hybrid --format json
```

Current shipped boundary:

- `--apply` is implemented for `github` and `hybrid`
- local promotion apply is not implemented yet
- repo-backed local promotion still uses the legacy `plan epic promote`
  compatibility path to create the local spec file today
- the promoted issue body becomes the canonical distilled planning artifact
- the original GitHub Discussion stays linked as collaboration history
- promotions with 5+ specs require `--project-decision create|skip` so project
  tracking is never silently skipped
- if a valid apply path fails on the GitHub API, Plan emits
  `manual_fallback_allowed=true`; only then may an agent use manual `gh`
  commands, followed by `plan github adopt`

When a multi-spec promotion is applied, `plan` will:

- create the initiative issue
- create the spec issues immediately
- create the milestone
- wire the initiative as parent of the spec issues
- add `blocked by` relationships only where the dependency plan says they are
  real
- mirror the created issue/milestone metadata into `.plan/.meta/github.json`

By default, new spec issues are `ready`. A spec starts as `needs-refinement`
only when the draft identified a concrete execution gap.

Recover manually-created or pre-existing planning issues with:

```bash
plan github adopt --project . --discussion 49 --issues 101,102,103 --format json
```

### 8. Preview Collaboration Guide Packets

`plan guide current` remains the guided brainstorm packet entry point:

```bash
plan guide current --project . --format json
```

Use `plan guide show` when you want an explicit preview of either:

- a brainstorm checkpoint from a guided session chain
- a collaboration stage driven by a brainstorm or GitHub Discussion source

Examples:

```bash
plan guide show --project . \
  --chain brainstorm/newsletter-system \
  --stage brainstorm \
  --checkpoint clarify-open-approaches \
  --format json

plan guide show --project . \
  --brainstorm newsletter-system \
  --stage promotion_review \
  --format json

plan guide show --project . \
  --discussion 49 \
  --stage initiative_draft \
  --format json
```

Current collaboration packet stages:

- `discussion_assess`
- `promotion_review`
- `initiative_draft`
- `spec_draft`
- `needs_refinement`

Current collaboration packet behavior:

- embeds the canonical `maturity_assessment` and `promotion_draft` payloads
- includes rendered initiative/spec draft markdown when the stage needs it
- emits explicit review and confirmation action objects for agent runtimes
- keeps JSON as the canonical output format
- does not mutate the source material while rendering the packet

### 9. Work The Spec

After promotion, the canonical execution contract may live in different places
depending on ownership:

- `local`: the repo-backed spec file under `.plan/specs/`
- `github`: the promoted GitHub spec issue body during shaping
- `hybrid`: the configured split between issue-backed planning and repo-backed
  spec material

Current shipped execution loop is still strongest on repo-backed specs. The new
GitHub collaboration foundation shapes and promotes work cleanly, but it does
not yet replace the local `plan spec ...` execution commands end-to-end.

For repo-backed specs, show the canonical spec:

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
- optional execution notes or planned slices

### 9. Analyze The Spec

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

### 10. Run A Spec Checklist

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

### 11. Approve The Spec

Execution should only start from an approved spec.

Approve it:

```bash
plan spec status --project . newsletter-system --set approved
```

Current spec statuses:

- `draft`
- `approved`
- `implementing`
- `done`

### 12. Spec Queue Workflow

Preferred execution loop is spec-first, not issue-first:

1. Establish queue with `plan status --project .`
2. Take the next approved spec from the queue
3. Run `plan spec execute --project . <spec-slug>` to start execution
4. Implement one slice
5. Review and verify that slice before committing it
6. Repeat slice-by-slice until the spec is done
7. Move to the next spec in queue if more queued specs remain
8. Open one PR when the queued specs for the branch are complete

Recommended commands:

```bash
plan status --project .

# take next approved spec
plan spec execute --project . <spec-slug>

# implement one slice
# run review + verification
# commit slice

# repeat until the spec is done
# then move to the next queued spec
```

Use this model:

- `plan status` = queue view
- approved spec = execution batch
- runtime slices = execution units inside the current spec
- review + verification happen before each slice commit
- PR = review for the completed queued spec batch

If GitHub integration is enabled, reconcile after merge:

```bash
git switch <integration-branch>
git pull --ff-only origin <integration-branch>
plan update --project .
plan github reconcile --project . --update-visible
plan status --project .
```

If GitHub integration is not enabled, use the same refresh without reconcile:

```bash
git switch <integration-branch>
git pull --ff-only origin <integration-branch>
plan update --project .
plan status --project .
```

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

## End-To-End Examples

```bash
plan init --project .
plan source set --project . local

plan brainstorm start --project . "Billing export"
plan brainstorm idea --project . billing-export --body "Export billing data to an external API"
plan brainstorm refine --project . billing-export
plan brainstorm challenge --project . billing-export

plan discuss assess --project . --brainstorm billing-export --format json
plan discuss promote --project . --brainstorm billing-export --format json
plan epic promote --project . billing-export

plan spec show --project . billing-export
plan spec analyze --project . billing-export
plan spec checklist --project . billing-export --profile api-integration
plan spec status --project . billing-export --set approved

plan spec execute --project . billing-export

plan status --project .
plan check --project .
```

GitHub collaborative example:

```bash
plan source set --project . github
plan discuss assess --project . --discussion 49 --format json
plan discuss promote --project . --discussion 49 --format json
plan discuss promote --project . --discussion 49 --apply --confirm --target github --format json
```

## What `plan` Does Not Do Right Now

Current state means:

- no memory or context-management features
- no fully finished cloud backend beyond the current GitHub foundation
- no end-to-end GitHub-native spec execution loop yet

Those are roadmap questions, not current usage.

## Practical Rules

- Start with the smallest useful pass.
- Use `refine` when the idea is fuzzy.
- Use `challenge` when the idea is too comfortable or too broad.
- Use `discuss assess` when you need an explicit maturity decision.
- Use `discuss promote` when you need a concrete promotion draft before writing to GitHub.
- Use `analyze` for general spec pressure-testing.
- Use `checklist` when the spec has domain-specific risk.
- Approve the spec before starting execution.
- Keep runtime slices small, concrete, and verification-aware.

## Current Command Surface

Top-level commands available today:

- `plan init`
- `plan adopt`
- `plan doctor`
- `plan update`
- `plan source`
- `plan brainstorm`
- `plan discuss`
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
plan source --help
plan brainstorm --help
plan discuss --help
plan epic --help
plan spec --help
plan story --help
plan github --help
```
