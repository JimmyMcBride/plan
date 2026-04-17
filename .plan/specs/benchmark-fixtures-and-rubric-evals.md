---
created_at: "2026-04-17T20:19:38Z"
epic: benchmark-fixtures-and-rubric-evals
project: plan
slug: benchmark-fixtures-and-rubric-evals
status: done
target_version: v5
title: Benchmark Fixtures and Rubric Evals Spec
type: spec
updated_at: "2026-04-17T20:34:02Z"
---

# Benchmark Fixtures and Rubric Evals Spec

Created: 2026-04-17T20:19:38Z

## Why

The roadmap reset says new shaping and prompt work should prove quality gains
before more complexity is added.

## Problem

Without benchmark fixtures and a rubric harness, `plan` has no durable way to
measure whether new guidance or workflows improve clarity and executability.

## Goals

- add realistic planning benchmark fixtures to the repo
- define rubric categories that match the product thesis
- provide a deterministic local maintainer evaluation harness
- document how to run and interpret the evaluation loop

## Non-Goals

- hosted eval infrastructure
- auto-tuning prompts via remote services
- scoring features visible to end users

## Constraints

- fixtures should be easy to inspect and update in git
- rubric categories should stay stable across minor iterations
- evaluation output should be deterministic and scriptable
- the harness must not require network access

## Solution Shape

- add benchmark fixtures under repo-owned testdata
- implement a small rubric/eval package that scores fixtures and responses
- expose the workflow through tests and maintainer docs first
- keep the harness simple enough to extend in later versions

## Flows

1. Maintainer runs the eval workflow locally or in CI.
2. The harness loads fixtures and rubric expectations.
3. Results show pass/fail or score deltas for the planned quality categories.
4. Maintainer uses the results to judge whether a new pass should stick.

## Data / Interfaces

- benchmark fixture files in `testdata/`
- rubric/eval code under `internal/`
- maintainer docs for running the benchmark loop

## Risks / Open Questions

- how much of the future scoring loop should be implemented now versus staged
  behind helpers
- how to keep fixtures realistic without creating heavy maintenance cost

## Rollout

- land the fixture format and baseline rubric harness in `v5`
- use tests as the first maintainer entry point
- only widen the interface after the local workflow proves useful

## Verification

- fixtures load deterministically in tests
- rubric scores are reproducible for the same input
- maintainer docs describe the evaluation loop accurately

## Story Breakdown

- [x] [Add planning benchmark fixtures](../stories/add-planning-benchmark-fixtures.md)
- [x] [Implement rubric evaluation harness](../stories/implement-rubric-evaluation-harness.md)
- [x] [Document and verify the maintainer eval workflow](../stories/document-and-verify-the-maintainer-eval-workflow.md)

## Analysis

### Missing Constraints

### Success Criteria Gaps

### Hidden Dependencies

### Risk Gaps

### What/Why vs How Leakage

### Recommended Revisions

## Resources

- [Epic](../epics/benchmark-fixtures-and-rubric-evals.md)
- [Product Direction](../PRODUCT.md)

## Notes
