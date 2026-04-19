---
created_at: "2026-04-18T22:26:06Z"
project: plan
slug: github-issues-integration
status: active
title: GitHub issues integration
type: brainstorm
updated_at: "2026-04-19T00:29:01Z"
---

# Brainstorm: GitHub issues integration

Started: 2026-04-18T22:26:06Z

## Focus Question

How should plan project local planning artifacts to GitHub issues without making GitHub the source of truth?
## Desired Outcome

Keep brainstorms, epics, and specs local and canonical inside `.plan`, while execution work lives in GitHub issues with clear blockers, ready order, async-safe parallel lanes, and links back to the local docs.

## Constraints

- Brainstorms, epics, and specs stay local in `.plan/`.
- If GitHub execution mode is enabled, stories should not be duplicated as first-class local markdown and GitHub issues at the same time.
- GitHub issues must link to canonical local epic/spec markdown files in the repo.
- Dependencies and blockers must make next-ready work obvious for humans and agents.
- plan remains planning-first and local-first for shaping; GitHub becomes the execution surface.

## Open Questions

- How should dependencies be encoded in GitHub issues: native GitHub features when available, issue body sections, or a plan-owned metadata convention?
- Should `plan` keep only minimal local issue metadata under `.plan/.meta/`, or some richer local index for readiness/status?
- Should closing an issue or merging a PR automatically advance the next-ready work, or only make it visible?
- Should GitHub execution mode be opt-in per repo or per epic/spec?
- When, if ever, should milestones enter the model?
## Ideas

- Use one GitHub issue per execution-ready story, with links to the canonical epic/spec markdown files.
- Render acceptance criteria, verification, blockers, and async notes directly into the issue body.
- Compute ready order from dependencies so the next executable issue is obvious.
- Allow multiple ready issues at once when blockers do not overlap, so async human/agent work can proceed safely.
- Store only issue metadata locally, not duplicate story notes.
- Optionally group issues under milestones later, not in the first cut.
## Raw Notes

## Refinement

### Problem

Teams using `plan` need local shaping depth and GitHub-native execution at the same time. Mirroring stories in both `.plan` and GitHub creates duplicated execution truth, weakens clarity, and makes it unclear where humans or agents should actually work from.

### User / Value

Indie developers and small teams keep local-first planning for shaping, while humans and AI agents get GitHub-native execution units with clear ordering, blockers, async-safe parallel work, and direct links back to the local epic/spec docs.

### Appetite

GitHub must not become source of truth for shaping. First cut should stay GitHub-issues-only, focused on issue-backed stories, blockers, ready order, and async flow. No full tracker parity. No duplicated local story notes.

### Remaining Open Questions

- Should the first release support issue creation only, or creation plus update/reconciliation?
- Should local status derive from issue state, or only surface issue state alongside epic/spec progress?
- How much dependency logic can rely on native GitHub features versus issue body conventions?
- Should milestone mapping wait until after issue-backed story flow proves useful?

### Candidate Approaches

- GitHub issue-backed stories: approved execution units materialize as GitHub issues, while `.plan` keeps only epic/spec/brainstorm docs plus issue metadata.
- Read-only execution backfill: local status views read issue and PR state without letting remote edits rewrite canonical shaping docs.
- Optional local-only mode remains for repos that do not enable GitHub execution.

### Decision Snapshot

Keep `.plan` canonical for brainstorms, epics, and specs. If GitHub integration is enabled, stories live in GitHub issues, not duplicate local markdown files. Start with explicit issue creation/update plus dependency-driven ready visibility.

## Challenge

### Rabbit Holes

- Duplicating story truth in both local markdown and GitHub issues.
- Trying to map every plan artifact type into GitHub.
- Building automation for labels, milestones, boards, projects, sub-issues, and realtime sync in the first cut.
- Depending too hard on GitHub-specific surfaces before the issue-backed execution model proves itself.

### No-Gos

- Do not make GitHub canonical.
- Do not duplicate canonical stories in both `.plan` and GitHub issues.
- Do not export brainstorms, epics, or specs as first-class GitHub issues.
- Do not support Jira/Linear-class tracker parity in the first cut.
- Do not silently rewrite local planning notes from remote issue edits.
- Do not create dependency logic that is only visible in `plan` and invisible from the issue itself.

### Assumptions

- Users already work in repos with GitHub Issues enabled.
- Story titles, acceptance criteria, verification, and blockers can be rendered into useful issue bodies.
- An issue can safely serve as the durable execution unit for a story.
- Local epic/spec markdown links are useful inside issue bodies.
- Explicit create/update commands are acceptable UX for v1.

### Likely Overengineering

Trying to support two-way field sync, issue templates, milestone planning, project-board automation, PR automation, and multi-tracker abstractions at the same time. Most likely failure mode: building an integration platform before proving the issue-backed execution loop.

### Simpler Alternative

Use GitHub issues as story storage, render blockers/spec links directly into the issue body, keep only minimal issue metadata locally, and compute ready work from dependencies without attempting full remote-local sync.
