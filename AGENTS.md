---
updated: "2026-09-01T14:51:24Z"
---
# Project Agent Contract

<!-- brain:begin agents-contract -->
Use this file as a Brain-managed project context entrypoint for `plan`.

Brain is intended for AI agents operating in this repo, not as a human-operated project dashboard.

Read the linked context files before substantial work. Prefer the `brain` skill and `brain` CLI for project memory, retrieval, and durable context updates.

## Table Of Contents

- [Overview](./.brain/context/overview.md)
- [Architecture](./.brain/context/architecture.md)
- [Standards](./.brain/context/standards.md)
- [Workflows](./.brain/context/workflows.md)
- [Memory Policy](./.brain/context/memory-policy.md)
- [Current State](./.brain/context/current-state.md)
- [Policy](./.brain/policy.yaml)

## Project Docs

- [README.md](./README.md)
- [gitflow.md](./docs/gitflow.md)
- [project-architecture.md](./docs/project-architecture.md)
- [project-overview.md](./docs/project-overview.md)
- [project-workflows.md](./docs/project-workflows.md)
- [using-plan.md](./docs/using-plan.md)

## Required Workflow

1. If no validated session is active, run `brain prep --task "<task>"`.
2. If a session is already active, run `brain prep`.
3. Read this file and the linked context files still needed for the task.
4. Use `brain context compile --task "<task>"` only when you need the lower-level packet compiler directly.
5. Retrieve project memory with `brain find plan` or `brain search "plan <task>"` when the compiled packet is not enough.
6. Use `brain edit` for durable context updates to AGENTS.md, docs, or .brain notes.
7. Run `brain context audit` after meaningful architecture, config, CI, deploy, test, or docs-surface changes.
8. Use `brain session run -- <command>` for required verification commands.
9. Finish with `brain session finish` so policy checks can enforce verification and surface promotion review when durable follow-through is still needed.

## Karpathy Guidelines

Behavioral guidelines to reduce common LLM coding mistakes, derived from [Andrej Karpathy's observations](https://x.com/karpathy/status/2015883857489522876) on LLM coding pitfalls.

Use these guidelines when writing, reviewing, or refactoring code to avoid overcomplication, make surgical changes, surface assumptions, and define verifiable success criteria.

**Tradeoff:** These guidelines bias toward caution over speed. For trivial tasks, use judgment.

### 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them; don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

### 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No flexibility or configurability that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

### 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't improve adjacent code, comments, or formatting unless the task requires it.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it; don't delete it.

When your changes create orphans:
- Remove imports, variables, and functions that your changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

### 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" -> "Write tests for invalid inputs, then make them pass"
- "Fix the bug" -> "Write a test that reproduces it, then make it pass"
- "Refactor X" -> "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:

```text
1. [Step] -> verify: [check]
2. [Step] -> verify: [check]
3. [Step] -> verify: [check]
```

Strong success criteria let you loop independently. Weak criteria such as "make it work" require constant clarification.

## Post-Adoption Enrichment

After `brain adopt` creates starter context, the AI agent must scan the repo before treating the templates as complete memory.

1. Treat generated context as starter context, not complete repo memory.
2. Scan repo structure, docs, manifests, entrypoints, tests, CI, config, and deployment surfaces.
3. Update AGENTS.md, docs, or .brain notes with durable project-specific findings.
4. Add focused .brain/resources notes for architecture, workflows, risks, and references that do not belong in top-level templates.
5. Keep generated managed blocks refreshable; put hand-authored findings in Local Notes or dedicated notes.
<!-- brain:end agents-contract -->

## Local Notes

Add repo-specific notes here. `brain context refresh` preserves content outside managed blocks.

### Repo Contract

- `plan` owns planning under `.plan/`.
- Do not store planning artifacts in Brain.
- Use `develop` as the default PR target for routine work.
- Never push directly to protected branches: `develop`, `release/*`, `main`.
- Linear source mode stores MVP integration identity in `.plan/.meta/linear.json`; Linear promotion requires `team_id` or `team_key` and remains agent/MCP-mediated unless a future spec adds direct API ownership.

### Karpathy Guidelines

Use these behavioral guidelines when writing, reviewing, or refactoring code to
reduce common LLM coding mistakes. They are derived from
[Andrej Karpathy's observations](https://x.com/karpathy/status/2015883857489522876)
on LLM coding pitfalls.

Tradeoff: these guidelines bias toward caution over speed. For trivial tasks,
use judgment.

#### Think Before Coding

Do not assume or hide confusion. Surface tradeoffs before implementing:

- State assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them instead of picking silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop, name what is confusing, and ask.

#### Simplicity First

Write the minimum code that solves the problem. Nothing speculative:

- No features beyond what was asked.
- No abstractions for single-use code.
- No flexibility or configurability that was not requested.
- No error handling for impossible scenarios.
- If 200 lines could be 50, rewrite it.

Ask: would a senior engineer say this is overcomplicated? If yes, simplify.

#### Surgical Changes

Touch only what is required. Clean up only your own mess:

- Do not improve adjacent code, comments, or formatting.
- Do not refactor things that are not broken.
- Match existing style, even if you would do it differently.
- If you notice unrelated dead code, mention it instead of deleting it.
- Remove imports, variables, or functions that your changes made unused.
- Do not remove pre-existing dead code unless asked.

Every changed line should trace directly to the user's request.

#### Goal-Driven Execution

Turn tasks into verifiable goals and loop until verified:

- "Add validation" means write tests for invalid inputs, then make them pass.
- "Fix the bug" means write a test that reproduces it, then make it pass.
- "Refactor X" means ensure tests pass before and after.

For multi-step tasks, state a brief plan:

```text
1. [Step] -> verify: [check]
2. [Step] -> verify: [check]
3. [Step] -> verify: [check]
```

Strong success criteria allow independent progress. Weak criteria such as
"make it work" require clarification.

### Optional Brain Setup

For Codex cloud environments, run:

```bash
./scripts/setup-codex-cloud.sh
```

That setup installs:

- repo-local Brain binary at `.codex/bin/brain`
- repo-local Brain skill at `.codex/skills/brain`

### Brain Usage Rules

- If `.codex/bin/brain` exists and the repo has a `.brain/` workspace, use Brain for context retrieval, prep, and session hygiene.
- If `.brain/` does not exist, skip Brain workflows. Do not create or adopt a Brain workspace unless the task explicitly asks for it.
- Keep `plan` as the source of truth for roadmap, brainstorms, epics, specs, and stories.

### Useful Commands

```bash
go run . update --project .
go run . check --project .
./scripts/refresh-plan-develop-context.sh
```

### Project Workflow Override

- Build from approved specs.
- Slice the current spec into execution-ready stories before coding.
- Finish one slice, review it, verify it, then commit that slice.
- Repeat until the current spec is complete.
- Move to the next queued spec only after the current spec is done.
- Before a PR is marked ready or merged, run `brain session finish`; if it
  requires durable notes, commit those notes on the same branch and retry finish.
- Open one PR after the queued specs for the branch are complete.

### Brain Planning Compatibility

- Compatible schema-v3 local planning commands delegate to the pinned stable
  Brain Planning packages; standalone Plan remains the CLI presentation host.
- Keep GitHub, hybrid, legacy, and unsupported paths on their existing
  standalone implementations until their migration slices are approved.
- Never use a local `replace` directive for the Brain dependency.
