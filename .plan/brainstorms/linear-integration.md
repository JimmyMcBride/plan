---
created_at: "2026-05-14T02:45:24Z"
project: plan
slug: linear-integration
status: active
title: Linear integration
type: brainstorm
updated_at: "2026-05-17T01:42:31Z"
---

# Brainstorm: Linear integration

Started: 2026-05-14T02:45:24Z

## Focus Question

## Desired Outcome

## Vision

Users should be able to brainstorm locally and ethereally until they have a mature initiative worth moving into Linear.

Linear mode should upload the initiative to Linear, break it into specs, and make the initiative plus child specs consumable by collaborators and agents.

The product shape should closely mirror how Plan's GitHub integration works, but Linear becomes the durable source of truth. Turning on Linear mode should expect the user to have the Linear MCP connected.

Linear integration is collaboration-first. Solo developers already have local and GitHub modes, so Linear should focus on teams and shared execution.

## Supporting Material

- None provided yet.

## Constraints

- First pass stays intentionally small: Plan promotes a mature local brainstorm/initiative into one Linear Project and a set of Linear Issues, with a required Linear team selected before issue creation. Linear Initiatives, status sync, reconcile, and full GitHub parity are deferred.

## Open Questions

- Does Plan own a direct Linear client, or does the agent use Linear MCP from a Plan-emitted promotion packet?
- How should Linear source mode coexist with local default, GitHub mode, and hybrid mode?
- What local mirror metadata is enough for traceability without duplicating Linear as source of truth?
## Ideas
## Raw Notes

Linear research notes:
- Linear Initiatives group Projects by company objective and are workspace-wide, long-horizon planning objects. Good later mapping for Plan phase/roadmap/program, not first-pass Plan initiative.
- Linear Projects are features or large units of work with clear outcomes, issues, optional documents, resources, milestones, progress graph, statuses, and multi-team support. Best first-pass mapping for Plan initiative.
- Linear Issues are required to belong to one team and carry status; they can belong to a Project. Best first-pass mapping for Plan specs.
- Linear Project milestones divide work inside a Project and can group issues by stage. Good optional mapping for rollout phases or spec groups, not required for basic promotion.
- Linear parent/sub-issues split work that is too large for a single issue but too small for a Project. Because Plan execution slices are ephemeral by default, sub-issues should be optional team-visible execution decomposition, not mandatory.
- Linear MCP gained initiative, project milestone, project update, and project label tools in 2026, so a Plan Linear backend can reasonably expect MCP-driven planning-object creation when the MCP is connected.
Recommendation: Plan initiative -> Linear Project; Plan spec -> Linear Issue in that project; Plan slices -> no persisted Linear object by default; optional sub-issues only when a spec needs visible team decomposition.

Handoff decisions:
- Use a 3-spec initiative for first implementation. Split: source/config model; Linear promotion packet plus metadata; agent/MCP docs and tests.
- MVP implementation contract: Plan does not call Linear directly. Plan emits deterministic `linear_promotion_packet`; the AI agent uses the connected Linear MCP to create Linear objects and reports created IDs back to Plan.
- Next artifact path: run challenge pass first, then promote the brainstorm into a GitHub idea/initiative issue using the repo's current planning flow.

Promotion refinement decisions:
- Dependency chain should be linear: source/config model first, promotion packet/metadata second, docs/guide/tests third.
- Do not create GitHub issues from the generic promotion draft. Tighten initiative and spec bodies first.
- Keep this as a local brainstorm source for discovery, then promote to GitHub planning issues through Plan once the draft is tight.

Dependency hints for promotion parser:
- Linear promotion packet, MCP handoff, and metadata recording depends on Linear source mode and team configuration.
- Linear integration docs, guide packets, and validation tests depends on Linear promotion packet, MCP handoff, and metadata recording.
## Refinement

### Problem

Teams already using Linear need Plan-shaped work to land where their collaboration, assignment, status, and execution conversations already happen, instead of forcing work through local files or GitHub issues.

### User / Value

Collaborators can review, assign, discuss, and execute Plan-created work in Linear, while agents can consume Linear Project and Issue data through the connected Linear MCP. The value is less translation between planning and team execution.

### Appetite

Small MVP. Build the minimum useful Linear source-of-truth path first: configure Linear mode/team, promote local planning into Linear Project + Issues, and persist Linear IDs in Plan metadata so agents can find the canonical Linear objects.

### Remaining Open Questions

- Does Plan own a direct Linear client, or does the agent use Linear MCP from a Plan-emitted promotion packet?
- How should Linear source mode coexist with local default, GitHub mode, and hybrid mode?
- What local mirror metadata is enough for traceability without duplicating Linear as source of truth?

### Candidate Approaches

- Recommended MVP approach: add Linear as a first-class source mode, but use an agent-mediated MCP promotion flow first. Plan emits structured Linear promotion packets for Project + Issue creation, the AI agent uses the connected Linear MCP to perform Linear mutations, and Plan records minimal mirror metadata: Linear workspace/team, project ID/url, issue IDs/urls, source brainstorm slug, and promotion timestamp.

### Decision Snapshot

- First implementation should be a 3-spec initiative.
- Spec split should separate source/config model, Linear promotion packet plus metadata, and agent/MCP docs/tests.
- Plan should not call Linear directly in the MVP.
- MVP contract: Plan emits deterministic `linear_promotion_packet`; the AI agent uses connected Linear MCP to create Linear objects and reports created IDs back to Plan.
- Next artifact path: run challenge pass first, then promote this brainstorm into a GitHub idea/initiative issue through Plan's current planning flow.

## Challenge

### Rabbit Holes

- Direct Linear API/auth inside Plan is the biggest rabbit hole. It creates auth, token storage, API lifecycle, and provider-client surface before the MCP-first promotion contract is proven.

### No-Gos

- No direct Linear mutations from Plan CLI in MVP. Agent with connected Linear MCP owns Linear mutations; Plan owns deterministic packet generation and minimal local metadata recording.

### Assumptions

- Linear MCP can create or update Linear Projects and Issues reliably enough through agent flow, and a selected Linear team is enough configuration for first promotion.

### Likely Overengineering

The work is most likely to become overengineered if we build full source-of-truth parity, direct auth/API support, status sync, reconcile, Linear Initiatives, milestones, or sub-issue decomposition before proving Project + Issue promotion.

### Simpler Alternative

Packet-only POC: Plan emits a deterministic `linear_promotion_packet` for a mature brainstorm/initiative but does not create Linear objects yet.

## Specs

- Linear source mode and team configuration
- Linear promotion packet, MCP handoff, and metadata recording
- Linear integration docs, guide packets, and validation tests

## Promotion Draft Requirements

### Initiative: Linear integration

Purpose:
- Add Linear as a collaboration-first source-of-truth backend for teams that already run execution in Linear.
- Keep local brainstorming ephemeral until the work matures enough to promote.
- Promote a mature Plan initiative into one Linear Project and child Linear Issues for specs.
- Preserve Plan's local-first default and GitHub backend while adding Linear as another explicit backend choice.

Scope:
- Linear is a first-class source mode alongside local, github, and hybrid.
- MVP uses an agent-mediated MCP flow instead of direct Linear API calls.
- MVP requires a selected Linear team before promotion.
- MVP stores minimal mirror metadata linking Plan source material to created Linear Project and Issue IDs/URLs.

Non-goals:
- No direct Linear API/auth/client implementation in Plan CLI.
- No Linear Initiative creation in the first pass.
- No status sync, reconcile, or full GitHub parity.
- No persisted Linear sub-issues for Plan execution slices by default.

### Spec 1: Linear source mode and team configuration

Goal:
- Extend Plan's backend/source model so Linear can be selected intentionally without weakening local-first defaults.

Scope:
- Add `linear` as an explicit source-of-truth mode or provider value in the CLI/config model.
- Capture the minimum Linear configuration needed for MVP promotion, especially workspace/team identity.
- Validate that Linear promotion cannot proceed without a configured or selected team.
- Keep GitHub and local behavior unchanged.

Acceptance:
- `plan source show` or equivalent output can represent Linear mode/config without ambiguity.
- Linear mode clearly communicates that durable planning data will live in Linear after promotion.
- Missing team configuration fails early with actionable guidance.
- Existing local/github/hybrid tests continue to pass.

Dependencies:
- Blocked by nothing.

### Spec 2: Linear promotion packet, MCP handoff, and metadata recording

Goal:
- Define and emit the deterministic promotion contract that lets an AI agent create Linear objects through the connected Linear MCP.

Scope:
- Add a `linear_promotion_packet` shape for a mature local brainstorm/initiative.
- Map Plan initiative to Linear Project.
- Map Plan specs to Linear Issues in the selected Linear team and project.
- Include rendered Linear Project and Issue markdown payloads in the packet.
- Include explicit MCP/agent action instructions and confirmation requirements.
- Record minimal mirror metadata after the agent reports created Linear object IDs/URLs.

Acceptance:
- Packet output is deterministic JSON and can be tested without calling Linear.
- Packet includes enough data for an agent to create one Linear Project plus spec Issues.
- Metadata recording stores Linear workspace/team, project ID/url, issue IDs/urls, source brainstorm slug, and promotion timestamp.
- Plan does not directly mutate Linear in this MVP.

Dependencies:
- Blocked by Spec 1.

### Spec 3: Linear integration docs, guide packets, and validation tests

Goal:
- Make Linear promotion understandable and safe for users and agents.

Scope:
- Update README/docs/skill guidance to explain Linear mode, MCP prerequisite, and MVP limitations.
- Add guide packet coverage so agents know when to ask for team selection, emit promotion packets, and wait for confirmation.
- Add validation/check behavior for missing Linear config or incomplete Linear metadata where appropriate.
- Document deferred work: direct API auth, Linear Initiatives, status sync, reconcile, milestones, and sub-issues.

Acceptance:
- Docs clearly distinguish local, GitHub, hybrid, and Linear ownership.
- Agent guidance explains that Linear MCP owns mutations in MVP.
- Tests cover guide/packet/docs-facing behavior and validation failures.
- Existing GitHub collaboration docs remain accurate.

Dependencies:
- Blocked by Spec 2.
