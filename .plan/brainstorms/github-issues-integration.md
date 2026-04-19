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

Publish execution work to GitHub issues while keeping brainstorms, epics, and specs local and canonical inside .plan.

## Constraints

- GitHub issues should project stories or execution tasks, not replace local planning artifacts.
- plan remains planning-only and local-first; GitHub stays external execution surface.

## Open Questions

- Should issue body edits ever sync back into local story notes, or should sync stay one-way at first?
- Should merged PRs be allowed to update local story status automatically?
- Should issue publication happen only for selected stories, or for every approved story by default?
- When, if ever, should milestones enter the model?

- Should issue body edits ever sync back into local story notes, or should sync stay one-way at first?

- Should merged PRs be allowed to update local story status automatically?

- Should issue publication happen only for selected stories, or for every approved story by default?
## Ideas

- Export approved stories to issues with stable links back to the local story note.

- Optionally group issues under milestones mapped from roadmap phases later, not in the first cut.
## Raw Notes

## Refinement

### Problem

Teams using plan need a way to expose execution work in GitHub without moving planning authority out of the repo. Today the rich planning flow lives in .plan, but collaborators, PR links, and day-to-day execution often happen in GitHub issues.

### User / Value

Indie developers and small teams keep the local-first planning depth they like, while still getting shareable GitHub-native execution tickets for collaborators, reviews, and repo automation.

### Appetite

GitHub must not become source of truth. Brainstorms, epics, and specs stay local. First cut should stay GitHub-only, issue-focused, and smaller than full tracker sync. Publishing should be explicit and reversible in local notes.

### Remaining Open Questions

- Should the first release support issue creation only, or issue update/reconciliation too?
- Should local story status ever be derived from issue state, or only displayed alongside it?
- Should milestone mapping wait until after basic story-to-issue publishing proves useful?

### Candidate Approaches

- One-way story publish: approved or selected stories create/update linked GitHub issues, with local note metadata storing issue number and URL.
- Read-only execution backfill: pull issue and PR state into local status views without letting issue body edits rewrite canonical local notes.
- Selective publish model: let users choose which stories stay local-only and which ones project to GitHub.

### Decision Snapshot

Keep .plan canonical. GitHub issues represent stories only. Start with explicit one-way publish plus optional read-only status backfill later.

## Challenge

### Rabbit Holes

- Full bidirectional sync between issue bodies and local story notes.
- Trying to map every plan artifact type into GitHub.
- Automated label, milestone, project board, and sub-issue orchestration in the first cut.
- Webhook-driven realtime sync from day one.

### No-Gos

- Do not make GitHub canonical.
- Do not export brainstorms, epics, or specs as first-class GitHub issues.
- Do not support Jira/Linear-class tracker parity in the first cut.
- Do not silently rewrite local planning notes from remote issue edits.

### Assumptions

- Users already work in repos with GitHub Issues enabled.
- Story titles and acceptance criteria can be rendered into useful issue bodies.
- A local story can safely hold stable issue metadata and links.
- Explicit publish/update commands are acceptable UX for v1.

### Likely Overengineering

Trying to support two-way field sync, issue templates, milestone planning, PR automation, and multi-tracker abstractions at the same time. Most likely failure mode: building an integration platform before proving the one-way story projection loop.

### Simpler Alternative

Export approved stories to GitHub issues with stable local links, optional label mapping, and no back-sync except maybe read-only status inspection. Everything above stories stays local.
