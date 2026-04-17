---
created_at: "2026-04-17T20:19:38Z"
project: plan
slug: benchmark-fixtures-and-rubric-evals
spec: benchmark-fixtures-and-rubric-evals
title: Benchmark Fixtures and Rubric Evals
type: epic
updated_at: "2026-04-17T20:19:38Z"
---

# Benchmark Fixtures and Rubric Evals

Created: 2026-04-17T20:19:38Z

## Outcome

Add a maintainable benchmark and rubric workflow that lets `plan` evaluate
whether new shaping passes actually improve plan quality before more surface is
added.

## Why Now

The product reset explicitly says prompt and workflow changes should be proven,
not assumed. Without fixtures and rubrics, `v5` risks growing based on taste
instead of measured planning quality.

## Scope Boundary

- benchmark fixtures for realistic planning scenarios
- rubric categories aligned with the product thesis
- a deterministic local evaluation harness for maintainers
- docs and tests that keep the eval loop usable

Not in scope:

- remote eval services
- leaderboard or hosted dashboards
- user-facing scoring UIs

## Spec

- [Draft Spec](../specs/benchmark-fixtures-and-rubric-evals.md)

## Resources

- [Product Direction](../PRODUCT.md)
- [Roadmap](../ROADMAP.md)

## Progress

- Target version: `v5`
- Status: planned

## Notes

This epic protects the product from drifting back toward power-first feature
work without evidence that the shaping loop is improving.
