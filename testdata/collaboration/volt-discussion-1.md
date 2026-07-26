## Problem

AI coding agents must coordinate changes across declarations, callers, matches, effects, modules, and tests, but existing language tooling often exposes these obligations only after partial failure.

## Goals

- Evaluate safe evolution of existing codebases as a primary use case.
- Make required change impact explicit, deterministic, and machine-readable.
- Preserve unrelated behavior and expose incomplete repository-wide propagation.

## Non-goals

- Production-grade language or ecosystem work in v0.
- Automatic application of compiler repairs.

## Constraints

- GitHub Discussions, Issues, Milestones, and Projects remain canonical planning state.
- The v0 kernel stays deliberately small.

## Proposed Shape

Use a small language kernel, deterministic reference interpreter, machine-readable diagnostics, and a controlled repository-level study.

## Promotion map

Target milestone: **Volt v0 — Evidence-Ready Research Prototype**

### Spec 1 — Research evidence and evaluation protocol

Update the evidence matrix, hypotheses, twelve-task repository-evolution workload, strict primary outcomes, fairness controls, six-endpoint analysis, power procedure, operational metrics, schemas, traceability, and falsification thresholds.

Acceptance criteria:

- Prior evidence, counterevidence, Volt hypotheses, and implementation directions are distinguished.
- `repository_change_success_rate` has a reproducible all-criteria definition with no partial credit.
- Expected impact surfaces and preservation assertions are frozen before confirmatory execution.
- All eleven metrics have deterministic definitions and fixtures; no composite is authoritative.
- The six-comparison Holm family, power rule, support rule, and falsification rule are machine-readable.
- Safe evolution remains explicitly unvalidated.

Dependencies: none.

Approval note: this is a material amendment to existing Issue #3 and requires owner reapproval.

### Spec 2 — Volt v0 language kernel and canonical syntax

Specify grammar, static semantics, explicit imports and module boundaries, closed ADTs, exhaustiveness, `Result`, exact effect sets, accepted/rejected examples, canonical formatting, and feature boundaries. Add the deterministic impact-list rule for public type, contract, effect, and module-boundary changes without expanding the v0 feature set.

Acceptance criteria:

- Every required construct has one canonical spelling and parse shape.
- Type, effect, match, import, and public-contract obligations are deterministic.
- Public-change categories define stable affected-symbol and diagnostic obligations.
- Accepted and rejected fixtures cover every v0 construct, exclusion, and supported public-change category.
- The kernel supports all twelve maintenance tasks without deferred features.

Dependencies: blocked by Spec 1.

### Spec 3 — Reference interpreter, program graph, and DiagnosticV1 protocol

Specify the compiler pipeline, stable symbol identity, complete program-graph node/edge coverage, repository-level impact analysis, CLI behavior, deterministic runtime capabilities, versioned `DiagnosticV1`, text/NDJSON parity, declarative repair surfaces, semantic-diff facts, and conformance tests.

Acceptance criteria:

- Definitions, references, imports, callers, public contracts, ADT variants, matches, effects, operations, and related tests are represented and queryable.
- Public changes produce deterministic affected declarations, missing propagation sites, dependency reasons, and bounded repair surfaces.
- Impact results never report unrelated modules for covered compiler-level fixtures.
- `DiagnosticV1` remains backward-compatible, stable-ordered, source-located, and renderer-equivalent.
- Semantic diff distinguishes public-contract, effect, ADT, match-coverage, and unexpected-surface facts without a composite score.
- The interpreter runs benchmark capabilities deterministically without network access.

Dependencies: blocked by Specs 1 and 2.

### Spec 4 — Benchmark corpus and controlled agent study

Build twelve existing-repository tasks, hidden requested-behavior and preservation tests, expected-impact manifests, causal ablations, descriptive TypeScript/Rust/Gleam baselines, pinned model and prompt manifests, randomized execution, pilot and six-endpoint power analysis, raw-result storage, and complete reporting.

Acceptance criteria:

- The corpus contains three tasks in each maintenance family and every task begins from a passing repository.
- Every task freezes requested behavior, minimum repair surface, actual required propagation sites, preservation assertions, and mutation-checked hidden tests.
- Hidden tests detect incomplete matches, forgotten callers, stale contracts, effect mistakes, weakened invariants, unrelated changes, and guarantee bypass.
- The harness records first and final repository-change success plus every operational metric.
- Causal and descriptive results are separated and all positive, negative, inconclusive, and falsifying results are reported.
- Runs are reproducible from content-addressed pinned inputs.

Dependencies: blocked by Specs 1, 2, and 3.
