# AGENTS

## Repo Contract

- `plan` owns planning under `.plan/`.
- Do not store planning artifacts in Brain.
- Use `develop` as the default PR target for routine work.
- Never push directly to protected branches: `develop`, `release/*`, `main`.

## Optional Brain Setup

For Codex cloud environments, run:

```bash
./scripts/setup-codex-cloud.sh
```

That setup installs:

- repo-local Brain binary at `.codex/bin/brain`
- repo-local Brain skill at `.codex/skills/brain`

## Brain Usage Rules

- If `.codex/bin/brain` exists and the repo has a `.brain/` workspace, use Brain for context retrieval, prep, and session hygiene.
- If `.brain/` does not exist, skip Brain workflows. Do not create or adopt a Brain workspace unless the task explicitly asks for it.
- Keep `plan` as the source of truth for roadmap, brainstorms, epics, specs, and stories.

## Useful Commands

```bash
go run . update --project .
go run . check --project .
./scripts/refresh-plan-develop-context.sh
```

