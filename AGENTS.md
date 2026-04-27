# Project Agent Contract

<!-- brain:begin agents-contract -->
Use this file as a Brain-managed project context entrypoint for `plan`.

Read the linked context files before substantial work. Prefer the `brain` skill and `brain` CLI for project memory, retrieval, and durable context updates.

## Table Of Contents

- [Overview](./.brain/context/overview.md)
- [Architecture](./.brain/context/architecture.md)
- [Standards](./.brain/context/standards.md)
- [Workflows](./.brain/context/workflows.md)
- [Memory Policy](./.brain/context/memory-policy.md)
- [Current State](./.brain/context/current-state.md)
- [Policy](./.brain/policy.yaml)

## Human Docs

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
7. Use `brain session run -- <command>` for required verification commands.
8. Finish with `brain session finish` so policy checks can enforce verification and surface promotion review when durable follow-through is still needed.
<!-- brain:end agents-contract -->

## Local Notes

Add repo-specific notes here. `brain context refresh` preserves content outside managed blocks.

### Repo Contract

- `plan` owns planning under `.plan/`.
- Do not store planning artifacts in Brain.
- Use `develop` as the default PR target for routine work.
- Never push directly to protected branches: `develop`, `release/*`, `main`.

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
- Before opening or merging the PR, run `brain session finish`; if it requires
  durable notes, commit those notes on the same branch and retry finish.
- Open one PR after the queued specs for the branch are complete.
