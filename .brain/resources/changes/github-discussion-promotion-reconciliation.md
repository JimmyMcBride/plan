---
title: GitHub Discussion Promotion Reconciliation
updated: "2026-07-26T08:32:45Z"
---
## Changes

- GitHub Discussion promotion preview now reconciles existing Plan-managed initiatives, specs, milestones, parent/sub-issue relationships, and blocked-by dependencies before proposing creation.
- Identity precedence favors exact Plan metadata and Discussion source identity, then stable slugs and established parent/repository metadata; equal candidates block preview.
- Structured Promotion map briefs render per-spec purpose or problem, scope, acceptance criteria, verification, dependencies, and readiness instead of copying global Discussion sections.
- Preview remains read-only. Confirmed apply creates, updates, reuses, or leaves artifacts unchanged and skips existing relationships, making repeated apply idempotent.
- The Volt Discussion #1 fixture resolves initiative #2, milestone #1, specs #3 through #6, and the existing dependency graph without new planning artifacts.
