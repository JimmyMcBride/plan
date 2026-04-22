# Guide Packet Schema and CLI Design

Date: 2026-04-21

## Why

`plan` already knows live planning state: current chain, current stage, current
cluster, linked artifacts, session summary, and next action. That makes `plan`
the right runtime source of truth for agent guidance.

Installed skills still help as bootstrap, but static agent personas are the
wrong primary mechanism for guided planning because they do not know:

- current chain and stage
- active artifact path
- current session summary and next action
- story backend mode
- stale downstream review state

The right split is:

- `plan` owns live stage guidance
- agents ask `plan` for a guide packet
- installed `plan` skill becomes a thin bootstrap that teaches agents to call
  `plan guide ...`

## Design Goals

- keep `plan` planning-only
- do not call model APIs from `plan`
- make guide output machine-readable first
- preserve a human-readable rendering for debugging and copy/paste use
- reuse existing guided-session state instead of inventing parallel metadata
- keep the CLI surface small

## Non-Goals

- per-agent installed persona bundles as the primary runtime interface
- direct OpenAI, Anthropic, or other model invocation from `plan`
- hidden session mutation during guide rendering
- a second artifact hierarchy outside `.plan/`

## Terms

- guide packet: machine-readable stage contract emitted by `plan`
- stage: runtime planning step, with brainstorm shipped first and later stages
  extending toward initiative/spec/execution guidance
- checkpoint: stage-local step such as `vision-intake` or
  `clarify-constraints-appetite`
- family overlay: optional style layer such as `gpt_style` or
  `reasoning_heavy`

## Core Decision

The structured packet is canonical. Rendered prompt text is derived.

That means:

- agents should prefer structured packet fields when possible
- `rendered_prompt` exists for runtimes that still want one prompt string
- prompt wording can evolve without breaking the machine contract as long as
  the structured schema stays stable

## CLI Surface

### `plan guide current`

Purpose:
- return the guide packet for the last-active guided session

Command:

```bash
plan guide current --project . --format json
```

Behavior:

- reads `.plan/.meta/guided_sessions.json`
- resolves `last_active_chain`
- builds a packet from current stage, checkpoint, session summary, linked
  artifact, and workspace rules
- does not mutate session state

Flags:

- `--format json`
  - default: `json`

Errors:

- no active guided session
- active chain missing from session state

Exit codes:

- `0` success
- `2` usage or missing-session errors

### `plan guide show`

Purpose:
- return a guide packet for an explicit chain and stage

Command:

```bash
plan guide show \
  --project . \
  --chain brainstorm/guided-planning-system \
  --stage brainstorm \
  --checkpoint clarify-constraints-appetite \
  --format json
```

Behavior:

- reads the requested chain
- uses explicit checkpoint if provided
- falls back to session current stage/checkpoint when omitted
- currently supports brainstorm-stage preview only

Flags:

- `--chain <chain-id>`
  - required
- `--stage brainstorm`
  - optional today; any non-brainstorm value is rejected in v1
- `--checkpoint <label>`
  - optional
- `--format json`
  - default: `json`

Errors:

- unknown chain
- unsupported stage
- checkpoint incompatible with requested stage

Exit codes:

- `0` success
- `2` usage or lookup errors

### `plan guide schema`

Purpose:
- future follow-up for emitting the JSON Schema for the guide packet

Status:

- intentionally deferred from the shipped v1 slice

## Packet Schema V1

Canonical packet shape:

```json
{
  "schema_version": 1,
  "kind": "guide_packet",
  "generated_at": "2026-04-21T12:00:00Z",
  "builder": {
    "command": "plan guide current",
    "format": "json"
  },
  "workspace": {
    "project_root": "/home/jimmy/Projects/plan",
    "planning_mode": "guided",
    "story_backend": "github",
    "integration_branch": "develop"
  },
  "session": {
    "chain_id": "brainstorm/guided-planning-system",
    "current_stage": "brainstorm",
    "current_cluster": 3,
    "current_cluster_label": "clarify-constraints-appetite",
    "stage_statuses": {
      "brainstorm": "in_progress",
      "epic": "todo",
      "spec": "todo",
      "execution": "todo"
    },
    "summary": "Vision captured. Supporting material recorded.",
    "next_action": "Continue with open questions and candidate approaches."
  },
  "artifact": {
    "type": "brainstorm",
    "slug": "guided-planning-system",
    "title": "Guided planning system",
    "path": ".plan/brainstorms/guided-planning-system.md",
    "status": "active"
  },
  "mode": {
    "stage": "brainstorm",
    "checkpoint": "clarify-constraints-appetite",
    "pass": "brainstorm_refine"
  },
  "sources": [
    ".plan/PROJECT.md",
    ".plan/ROADMAP.md",
    ".plan/brainstorms/guided-planning-system.md",
    ".plan/specs/vision-intake-and-brainstorm-co-planning.md",
    ".plan/specs/guided-stage-handoffs-and-artifact-writing.md"
  ],
  "contract": {
    "role": "co_planning_facilitator",
    "stance": [
      "collaborative",
      "direct",
      "skeptical_when_needed",
      "keep_scope_small"
    ],
    "goal": "Turn raw vision into a promotable brainstorm note.",
    "question_strategy": {
      "cluster_size_min": 2,
      "cluster_size_max": 4,
      "reflect_once_per_cluster": true,
      "gap_guidance": "one_recommended_plus_up_to_two_alternatives",
      "menu_actions": [
        "continue",
        "refine",
        "stop_for_now"
      ]
    },
    "artifact_strategy": {
      "write_mode": "additive",
      "durable_artifact": ".plan/brainstorms/guided-planning-system.md",
      "strengthen_sections": [
        "Problem",
        "User / Value",
        "Constraints",
        "Appetite"
      ],
      "preserve_rules": [
        "User input first",
        "Do not draft spec content or execution slices during brainstorm stage"
      ]
    },
    "do": [
      "Ask only the current question cluster.",
      "Explain gaps when they appear.",
      "Push toward smaller, clearer scope boundaries."
    ],
    "avoid": [
      "Do not jump ahead into implementation details.",
      "Do not turn the interaction into a giant intake form.",
      "Do not silently rewrite the user intent."
    ],
    "quality_bar": [
      "Problem and user value are concrete.",
      "Scope boundary is visible.",
      "Open questions are blocker-shaped, not vague brainstorming sprawl."
    ],
    "completion_gate": [
      "The brainstorm can move into the next durable planning step without hidden assumptions.",
      "The recommended next stage is clear."
    ],
    "command_hints": [
      {
        "purpose": "resume_current_stage",
        "command": "plan brainstorm resume guided-planning-system --project ."
      },
      {
        "purpose": "move_to_next_stage",
        "command": "plan brainstorm resume guided-planning-system --project ."
      }
    ]
  },
  "rendered_prompt": "You are guiding the brainstorm stage for `plan`..."
}
```

## Field Contract

### Required Top-Level Fields

- `schema_version`
  - integer
  - packet schema version
- `kind`
  - string
  - always `guide_packet`
- `generated_at`
  - RFC3339 UTC timestamp
- `builder`
  - object
- `workspace`
  - object
- `mode`
  - object
- `contract`
  - object
- `rendered_prompt`
  - string

### Optional Top-Level Fields

- `session`
  - object
  - required for `guide current`
  - optional for explicit preview use cases later
- `artifact`
  - object
  - omitted only if no durable artifact exists yet
- `sources`
  - string array

### `builder`

- `command`
  - source command such as `plan guide current`
- `format`
  - currently `json`

### `workspace`

- `project_root`
  - absolute project path
- `planning_mode`
  - current planning mode such as `guided`
- `story_backend`
  - `local` or `github`
- `integration_branch`
  - current integration branch name when available

### `session`

Derived from existing guided-session state.

- `chain_id`
- `current_stage`
- `current_cluster`
- `current_cluster_label`
- `stage_statuses`
- `summary`
- `next_action`

This should map cleanly onto current fields in
`.plan/.meta/guided_sessions.json`.

### `artifact`

- `type`
  - `brainstorm`, `epic`, `spec`, or `story_set`
- `slug`
- `title`
- `path`
  - repo-relative path when local
- `status`

### `mode`

- `stage`
  - `brainstorm` today; later follow-up work may add initiative/spec/execution
    stages
- `checkpoint`
  - stage-local checkpoint label
- `pass`
  - one of:
    - `brainstorm_start`
    - `brainstorm_refine`
    - `brainstorm_challenge`
    - `brainstorm_intake`
    - `brainstorm_refine`
    - `brainstorm_handoff`

### `contract`

- `role`
  - concise machine-friendly role id
- `stance`
  - array of behavioral anchors
- `goal`
  - single sentence stage goal
- `question_strategy`
  - interaction contract
- `artifact_strategy`
  - artifact writing contract
- `do`
  - required behavior list
- `avoid`
  - forbidden or discouraged behavior list
- `quality_bar`
  - stage quality expectations
- `completion_gate`
  - what must be true before leaving stage
- `command_hints`
  - exact follow-up commands the agent can run

### `rendered_prompt`

Derived string rendering of the structured fields above.

Rules:

- machine contract stays authoritative
- rendering may change wording
- meaning must stay semantically equivalent

## Initial Checkpoint Vocabulary

Exact checkpoint labels should be reused from guided session and stage logic
where they already exist.

For brainstorm v1:

- `vision-intake`
- `clarify-problem-user-value`
- `clarify-constraints-appetite`
- `clarify-open-approaches`
- `handoff-epic`
  - legacy checkpoint id retained for compatibility; semantically this is the
    handoff where guide packet guidance decides whether the work should stay one
    bounded spec or split into multiple specs under one initiative

For later stages:

- use stage-local labels
- keep labels stable once emitted in packets
- prefer descriptive ids over display text

## Markdown Rendering Contract

Deferred follow-up, not part of the shipped v1 slice.

`--format md` should render:

1. stage and checkpoint
2. current artifact
3. current summary and next action
4. role and goal
5. do / avoid
6. quality bar
7. completion gate
8. command hints
9. rendered prompt block

This keeps one human-debug view without inventing a second schema.

## Text Rendering Contract

Deferred follow-up, not part of the shipped v1 slice.

`--format text` should be concise and terminal-friendly:

- one-line stage summary
- one-line artifact pointer
- short behavior bullets
- short next-command hints

No prose banner. No explanatory fluff. Deterministic ordering.

## Example Commands

Agent bootstrap flow:

```bash
plan brainstorm start --project . "Billing export"
plan guide current --project . --format json
```

Resume flow:

```bash
plan brainstorm resume billing-export --project .
plan guide current --project . --format json
```

Explicit stage preview:

```bash
plan guide show \
  --project . \
  --chain brainstorm/billing-export \
  --stage brainstorm \
  --checkpoint clarify-open-approaches \
  --format json
```

Deferred schema export:

```bash
plan guide schema --format json
```

## Error Contract

JSON mode:

- emit error text to stderr
- emit no partial JSON to stdout
- return exit code `2`

Examples:

- no active session:
  - `No active guided session. Start one with 'plan brainstorm start --project . "<topic>"'.`
- unknown chain:
  - `Guided session "brainstorm/foo" not found.`
- unsupported stage:
  - `Unsupported guide stage "release". Expected brainstorm.`

## Integration With Existing Skill

The installed `plan` skill should shrink to a bootstrap contract:

- detect `.plan/`
- call `plan guide current --project . --format json` after guided stage entry
- follow returned contract
- use `plan` commands for durable mutations

That keeps:

- installed skill small
- runtime guidance live
- agent coupling low
- model-family overlays optional instead of mandatory install targets

## Rollout Fit

This design fits existing v8 work:

- guided session engine provides packet input state
- vision intake and brainstorm co-planning provides first stage contract
- guided stage handoffs provide later-stage packet builders

Relevant specs:

- [Guided Session Engine and Resume](../specs/guided-session-engine-and-resume.md)
- [Vision Intake and Brainstorm Co-Planning](../specs/vision-intake-and-brainstorm-co-planning.md)
- [Guided Stage Handoffs and Artifact Writing](../specs/guided-stage-handoffs-and-artifact-writing.md)

## Recommended Next Slice

If this direction holds:

1. add `plan guide current|show|schema`
2. build brainstorm-stage packet first
3. decide whether model-family overlays are worth adding later
4. update installed `plan` skill to use guide packets instead of static stage prose
