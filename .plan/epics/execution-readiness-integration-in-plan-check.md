---
created_at: "2026-04-17T20:58:34Z"
project: plan
slug: execution-readiness-integration-in-plan-check
spec: execution-readiness-integration-in-plan-check
title: Execution-Readiness Integration in plan check
type: epic
updated_at: "2026-04-17T20:58:34Z"
---

# Execution-Readiness Integration in plan check

Created: 2026-04-17T20:58:34Z

## Outcome
Teach `plan check` to inspect the spec-to-story handoff, not just note sections
in isolation.

## Why Now
Current checks can pass even when an approved spec has no useful story
breakdown or when story coverage does not match the spec state. `v6` should
close that gap before bigger workflow power returns.

## Shape

### Appetite
One pass that adds cross-artifact readiness rules while keeping `plan check`
deterministic, local, and readable at project scale.

### Outcome
Maintainers can run `plan check` and quickly see whether a spec is actually
ready to become executable stories or whether the handoff still has holes.

### Scope Boundary
- spec-to-story readiness rules in `plan check`
- cross-artifact findings for approved and implementing specs
- deterministic project, epic, and spec scope reporting
- severity calibration between blocking errors and guidance warnings

### Out of Scope
- dependency scheduling or queue management
- mandatory critique before work can start
- external CI or tracker integrations

### Success Signal
`plan check` catches the most common v6 handoff failures before they turn into
weak or missing stories.

## Scope Boundary
- readiness findings for spec `## Story Breakdown`
- child-story coverage and status alignment
- command output that stays readable at larger scopes

Not in scope:

- project scheduling features
- automatic story creation
- remote enforcement

## Spec
- [Draft Spec](../specs/execution-readiness-integration-in-plan-check.md)

## Resources
- [Product Direction](../PRODUCT.md)
- [Roadmap](../ROADMAP.md)

## Progress
- Target version: `v6`
- Status: planned

## Notes
Checks should stay concrete and useful. Noise would undercut the whole `v6`
readiness push.
