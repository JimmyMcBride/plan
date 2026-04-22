---
created_at: "2026-04-21T06:36:20Z"
project: plan
slug: guide-packet-and-cli-foundation
status: approved
target_version: v8
title: Guide Packet and CLI Foundation Spec
type: spec
updated_at: "2026-04-21T06:39:20Z"
---

# Guide Packet and CLI Foundation Spec

Created: 2026-04-21T06:36:20Z

## Why

Guided planning quality should come from live `plan` state, not from static
persona text installed into each agent runtime.

## Problem

The agent-facing guidance for `plan` is still too static. An external agent can
learn the general workflow, but it cannot ask `plan` what guidance applies to
the current guided stage right now, so stage behavior drifts away from the live
planning state.

## Goals

- let an agent ask `plan` what guidance applies right now
- give the brainstorm stage one live source of truth for agent guidance
- support both current-session lookup and explicit chain lookup
- reduce reliance on stale installed prompt glue
- keep `plan` planning-only and model-free
- keep the first slice small enough to validate quickly

## Non-Goals

- direct model API calls from `plan`
- installed per-agent personas as the primary runtime interface
- full guide coverage for initiative, spec, or execution stages
- markdown or text guide rendering in the first slice
- automatic consumption by Codex, Claude, or other runtimes in the same scope
- schema export command in the first slice

## Constraints

- guide rendering must not mutate session state
- reuse `.plan/.meta/guided_sessions.json` as the session source of truth
- keep the CLI surface to `plan guide current` and `plan guide show`
- keep brainstorm checkpoint labels aligned with existing cluster ids
- packet output must be deterministic and versioned
- JSON mode must emit full packets to stdout and errors to stderr only
- actionable error text should exist when there is no active guided session
- the brainstorm note remains the durable artifact; guide packets are runtime
  contracts, not new planning documents
- the brainstorm handoff checkpoint must allow either a single-spec next step or
  a multi-spec initiative when the work is larger

## Solution Shape

- add a guide-packet builder that composes:
  - current workspace rules
  - current guided session state
  - linked brainstorm artifact metadata
  - brainstorm-stage contract data
- add `plan guide current --project . --format json`
- add `plan guide show --project . --chain ... --stage ... --checkpoint ... --format json`
- emit one machine-first packet with these top-level sections:
  - `schema_version`
  - `kind`
  - `generated_at`
  - `builder`
  - `workspace`
  - `session`
  - `artifact`
  - `mode`
  - `sources`
  - `contract`
  - `rendered_prompt`
- treat the structured packet as canonical and derive `rendered_prompt` from it
- support brainstorm-stage checkpoints:
  - `vision-intake`
  - `clarify-problem-user-value`
  - `clarify-constraints-appetite`
  - `clarify-open-approaches`
  - `handoff-epic` (legacy checkpoint id retained for compatibility while the
    decision itself now chooses between a single-spec path and a larger
    multi-spec initiative)

## Flows

1. User starts or resumes a guided brainstorm.
2. The agent calls `plan guide current --project . --format json`.
3. `plan` reads the last-active guided session and linked brainstorm note.
4. `plan` returns a JSON guide packet for the current brainstorm checkpoint.
5. The agent uses the packet contract to guide the next user interaction.
6. At the brainstorm handoff checkpoint, the packet guidance makes the
   single-spec vs multi-spec decision explicit based on the size of the work.
7. If the runtime needs an explicit preview instead of the active session, it
   calls `plan guide show --chain ... --stage brainstorm --checkpoint ...`.
8. `plan` returns the requested packet without mutating the session.

## Data / Interfaces

- new CLI group:
  - `plan guide current`
  - `plan guide show`
- packet schema v1:
  - `schema_version`: integer
  - `kind`: `guide_packet`
  - `generated_at`: RFC3339 UTC timestamp
  - `builder`: command metadata
  - `workspace`: project root, planning mode, story backend, integration branch
  - `session`: chain id, stage, checkpoint, summary, next action, statuses
  - `artifact`: type, slug, title, path, status
  - `mode`: stage, checkpoint, pass
  - `sources`: supporting local files used to shape the packet
  - `contract`: role, stance, goal, question strategy, artifact strategy, do,
    avoid, quality bar, completion gate, command hints
  - `rendered_prompt`: derived prompt string
- initial supported format:
  - `json`
- initial supported stage:
  - `brainstorm`

Reference design source:

- [Guide Packet Schema and CLI Design](../research/guide-packet-schema-and-cli-design.md)

## Risks / Open Questions

- whether family overlays belong in the first implementation slice or a follow-up
- whether `rendered_prompt` will tempt callers to ignore structured fields
- how much source context belongs in the packet before it becomes noisy
- whether `plan guide show` should require an explicit stage when the chain is
  already known
- the first slice should stay read-only; if packet work starts demanding stored
  metadata changes, that likely means the scope is too large
- no guided-session migration is planned in this slice, so follow-up work should
  prefer trimming packet fields over mutating durable session state

## Rollout

- land packet schema and builder types first
- ship brainstorm-stage packet generation plus `guide current`
- add `guide show` next
- validate the workflow manually with agent usage before changing installed
  skill behavior
- extend into later stages only after the brainstorm-stage packet proves useful
- keep rollout read-only against current guided-session state; if migration
  pressure appears, defer it to a separate follow-up spec

## Verification

- `plan guide current --format json` returns a valid packet when an active
  guided brainstorm session exists
- `plan guide current --format json` fails with a clear actionable error when
  no active guided session exists
- `plan guide show` returns a compatible packet for an explicit chain and
  checkpoint
- returned packets contain the current stage, checkpoint, summary, next action,
  and linked artifact path
- packet rendering does not rewrite `.plan/.meta/guided_sessions.json`
- `rendered_prompt` stays semantically aligned with the structured contract
- brainstorm checkpoint ids in the packet match current guided brainstorm
  cluster labels

## Implementation Slices

- Add guide packet types and brainstorm-stage packet builder
- Add `plan guide current` JSON command
- Add `plan guide show` JSON command

## Analysis

### Missing Constraints

- None.

### Success Criteria Gaps

- None.

### Hidden Dependencies

- None.

### Risk Gaps

- None.

### What/Why vs How Leakage

- [warn] The narrative sections include implementation detail that belongs in Solution Shape or Data / Interfaces.

### Recommended Revisions

- [warn] Keep ## Why, ## Problem, ## Goals, and ## Non-Goals product-facing, then move technical detail into ## Solution Shape or ## Data / Interfaces.

## Checklist

### general

status: ok
blocking_findings: 0
guidance_findings: 0

- [ok] No findings.
## Resources

- [Research: Guide Packet Schema and CLI Design](../research/guide-packet-schema-and-cli-design.md)
- [Guided Session Engine and Resume Spec](../specs/guided-session-engine-and-resume.md)
- [Vision Intake and Brainstorm Co-Planning Spec](../specs/vision-intake-and-brainstorm-co-planning.md)
- [Guided Stage Handoffs and Artifact Writing Spec](../specs/guided-stage-handoffs-and-artifact-writing.md)
- [Product Direction](../PRODUCT.md)

## Notes

Recommendation locked during planning: use installed skills as bootstrap only,
and make live guide packets the runtime source of truth for guided agent
behavior.

Current implementation choice: guide packet shipped as one bounded spec on one
branch. If the feature grows later, a multi-spec initiative remains valid, but
it is not required for the current slice.
