---
created_at: "2026-04-17T20:58:34Z"
epic: execution-readiness-integration-in-plan-check
project: plan
slug: execution-readiness-integration-in-plan-check
status: approved
target_version: v6
title: Execution-Readiness Integration in plan check Spec
type: spec
updated_at: "2026-04-17T20:58:34Z"
---

# Execution-Readiness Integration in plan check Spec

Created: 2026-04-17T20:58:34Z

## Why
`plan check` already validates spec and story sections. `v6` should extend that
work so the transition from approved spec to executable stories is checked too.

## Problem
The current checker can report a clean project even when an approved spec still
has a placeholder `## Story Breakdown`, no sliced stories, or story coverage
that does not match the spec's implementation state.

## Goals
- extend `plan check` with spec-to-story readiness rules
- detect approved specs that are not slice-ready or have thin `## Story
  Breakdown` guidance
- report cross-artifact gaps between spec status and linked story coverage
- keep output simple and consistent across project, epic, spec, and story
  scopes

## Non-Goals
- full project scheduling or dependency solving
- mandatory critique before any work begins
- remote sync or CI-only enforcement features

## Constraints
- checks must remain deterministic and local
- new findings must fit the existing `CheckFinding` model and formatter
- project-scope output must stay readable
- rules should distinguish blocking errors from guidance warnings

## Solution Shape
- add story-slicing readiness checks for approved and implementing specs
- compare spec status, `## Story Breakdown`, and child story presence together
- surface critique-aware or execution-ready guidance where available without
  requiring critique
- extend command and planning tests to cover the new findings

## Flows
1. User runs `plan check`.
2. `plan` loads specs, stories, and the spec-to-story relationships for the
   selected scope.
3. The report flags missing slice coverage, placeholder breakdown content, or
   status mismatches with clear fixes.
4. User strengthens the spec or stories before implementation continues.

## Data / Interfaces
- new `CheckFinding` rules for story-slice readiness and coverage
- helper logic that reads spec `## Story Breakdown` content alongside child
  stories
- existing `plan check` command output, extended with the new findings only

## Risks / Open Questions
- false positives on specs intentionally approved before slicing begins
- how aggressive blocking severity should be for missing story coverage
- whether project-scope output stays readable once cross-artifact findings grow

## Rollout
- land epic and spec scope readiness rules first
- calibrate severity in tests before widening to project scope
- only make issues blocking where the existing lifecycle already expects
  execution-ready stories

## Verification
- an approved spec with placeholder `## Story Breakdown` content produces
  readiness findings
- an implementing spec with missing or orphaned stories produces cross-artifact
  findings
- project and epic scope output remain deterministic and readable
- command and planning tests cover the new rules directly

## Story Breakdown
- [ ] [Add spec-to-story readiness rules in plan check](../stories/add-spec-to-story-readiness-rules-in-plan-check.md)
- [ ] [Add cross-artifact check coverage for v6 readiness](../stories/add-cross-artifact-check-coverage-for-v6-readiness.md)
- [ ] [Document v6 execution-readiness check workflow](../stories/document-v6-execution-readiness-check-workflow.md)

## Analysis

### Missing Constraints

### Success Criteria Gaps

### Hidden Dependencies

### Risk Gaps

### What/Why vs How Leakage

### Recommended Revisions

## Checklist

## Resources
- [Epic](../epics/execution-readiness-integration-in-plan-check.md)
- [Product Direction](../PRODUCT.md)

## Notes
