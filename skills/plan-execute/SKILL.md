---
name: plan-execute
description: Execute an approved Plan spec from spec slug or GitHub URL through implementation, checks, commit, push, and a ready pull request. Use this skill whenever the user says `$plan-execute`, asks to execute or implement a Plan spec, provides a Plan spec slug, asks for a spec-to-PR workflow, or asks to turn an approved planning issue/spec into code.
user-invocable: true
args:
  - name: spec
    description: Plan spec slug or GitHub issue/discussion URL to execute.
    required: true
---

# Plan Execute

Use this skill to move one approved Plan spec from planning truth to a pull
request. `plan` owns the spec contract; this skill owns the execution rail:
git, code, tests, screenshots, commits, pushes, PRs, and Brain session hygiene.

## Contract

- Resolve exactly one spec from a `.plan` slug or GitHub URL before coding.
- Verify the spec is approved and executable. Stop if it is vague, blocked,
  unsafe, or not approved.
- Keep implementation traceable from spec -> slices -> commits -> PR.
- Create ready PRs by default. Use draft PRs only when the user explicitly asks
  for draft mode.
- Do not run production migrations, production seed changes, or destructive data
  operations unless the user explicitly asks.
- Do not open a PR until relevant tests pass, or failures are clearly documented
  with the reason the PR is still needed.
- Capture screenshots or manual QA notes only when UI behavior changes.

## Startup

1. Read repo instructions such as `AGENTS.md`.
2. If a Brain workspace exists, start or reuse a Brain session for the task.
3. Read `.plan/PROJECT.md`, `.plan/ROADMAP.md`, and the target spec.
4. Check source mode with `plan source show --project .`.
5. If the spec lives in GitHub or source mode is `github`/`hybrid`, inspect the
   linked GitHub issue, discussion, milestone, and labels before coding.
6. Confirm the worktree is clean enough to safely branch. Do not overwrite user
   changes.

## Execution Flow

Use this sequence unless repo instructions require a stricter flow:

1. Resolve input:
   - local spec: `plan spec show --project . <spec-slug>`
   - GitHub URL/number: inspect with Plan/GitHub commands, then map to the Plan
     spec or planning issue.
2. Verify readiness:
   - spec status is approved or equivalent
   - acceptance criteria and verification expectations are clear
   - dependencies, migrations, and rollout constraints are understood
3. Sync integration branch:
   - use the repo's default development branch, commonly `develop`
   - fetch and fast-forward before branching
4. Create branch:
   - branch name: `codex/<spec-slug>`
   - if branch exists, inspect it before reusing or creating a suffixed branch
5. Derive slices:
   - run `plan spec execute --project . <spec-slug>`
   - treat generated slices as ephemeral execution order, not new planning truth
6. Implement slices in order:
   - finish one coherent slice before starting the next
   - run focused checks after each slice
   - update Plan/GitHub story or spec status only when the repo workflow expects it
7. Final review:
   - inspect full diff
   - run full relevant checks
   - run migrations/seeds only in non-production test/dev contexts when needed
   - capture screenshots/manual QA notes for UI work
8. Commit:
   - stage only intentional changes
   - use a clear message referencing the spec when useful
9. Push and PR:
   - push `codex/<spec-slug>`
   - open a ready PR to the repo's integration branch
   - include summary, spec link, slice list, tests run, migration/seed notes,
     manual QA, and screenshots when relevant
10. Finish:
   - finish the Brain session when one is active
   - report PR URL, tests, and any documented risk

## Stop Conditions

Stop and ask or repair planning state before coding when:

- the input does not resolve to exactly one spec
- the spec is not approved
- `plan spec execute` shows the work is too vague or blocked
- source mode or GitHub metadata disagrees with the requested execution target
- required credentials, services, or test data are missing
- implementation requires production data changes not explicitly authorized
- existing user changes make a safe branch or commit impossible

## PR Body

Use this structure unless the repo has a stricter template:

```markdown
## Summary

## Spec

## Slices

## Tests

## Migrations / Seeds

## Manual QA

## Screenshots
```

Write `None` for migration/seed, manual QA, or screenshot sections that do not
apply. If tests fail and a PR is still opened, name the failing command and the
reason it is not fixed in this PR.
