---
created_at: "2026-04-17T20:19:39Z"
epic: spec-checklist-profiles
project: plan
slug: spec-checklist-profiles
status: done
target_version: v5
title: Spec Checklist Profiles Spec
type: spec
updated_at: "2026-04-17T20:34:02Z"
---

# Spec Checklist Profiles Spec

Created: 2026-04-17T20:19:39Z

## Why

`plan spec analyze` provides a general diagnostic pass, but some planning work
needs more targeted questions.

## Problem

UI-heavy flows, integrations, and migrations each fail in different ways.
Without checklist profiles, the tool cannot offer deeper rigor where it matters
most.

## Goals

- add a durable `## Checklist` section to spec notes
- ship `plan spec checklist`
- support named profiles for general, UI flow, API integration, and data
  migration review
- keep the checklist pass additive and non-destructive

## Non-Goals

- domain registries loaded from the network
- automatic spec approval
- checklist modes for every possible specialty on day one

## Constraints

- checklist output should live in the spec note
- profiles must be deterministic and versionable in repo code
- reruns should replace or refresh only checklist output
- findings should distinguish blocking from advisory issues cleanly

## Solution Shape

- extend the spec template with a fixed `## Checklist` section
- add profile definitions in code
- implement a command that writes checklist results and returns a meaningful exit
  status
- keep profile output readable for both humans and agents

## Flows

1. User drafts or updates a spec.
2. User runs `plan spec checklist <epic-slug> --profile <profile>`.
3. `plan` evaluates the spec against the chosen profile.
4. Results are written under `## Checklist` and summarized in the CLI output.

## Data / Interfaces

- spec template additions
- checklist profile definitions in planning code
- `plan spec checklist` command flags and exit behavior

## Risks / Open Questions

- how aggressively the profile pass should block compared with `spec analyze`
- whether one spec should support multiple stored profile reports cleanly

## Rollout

- land the note schema and a small initial profile set in `v5`
- validate the results through tests and docs
- expand profiles only after the first set proves useful

## Verification

- specs include the `## Checklist` headings
- profile runs are deterministic and idempotent
- exit codes reflect blocking issues where the profile says they should

## Story Breakdown

- [x] [Add spec checklist note schema and profiles](../stories/add-spec-checklist-note-schema-and-profiles.md)
- [x] [Implement spec checklist command and findings](../stories/implement-spec-checklist-command-and-findings.md)
- [x] [Add spec checklist tests and docs](../stories/add-spec-checklist-tests-and-docs.md)

## Analysis

### Missing Constraints

### Success Criteria Gaps

### Hidden Dependencies

### Risk Gaps

### What/Why vs How Leakage

### Recommended Revisions

## Resources

- [Epic](../epics/spec-checklist-profiles.md)
- [Product Direction](../PRODUCT.md)

## Notes
