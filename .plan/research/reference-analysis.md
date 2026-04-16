# Reference Analysis

Date: 2026-04-15

## Snapshots

- `brain` at `36828bfdff7e` — `2026-04-15 Ship context compiler and automatic project upgrades (#10)`
- `beads` at `4e3dbb6db0c7` — `2026-04-15 fix(nix): update vendorHash after dependabot go.sum bumps (GH#3221) (#3297)`
- `get-shit-done` at `712e381f1393` — `2026-04-15 docs: document required Bash permission patterns for executor subagents (#2071) (#2288)`

## brain

### What works

- Local-first markdown artifacts inside repo.
- Default flow already strong: brainstorm -> epic -> spec -> stories.
- Canonical spec gate before stories. Good constraint. Prevents premature execution.
- Skill install path already solid. Local and global agent install. Easy mental model.
- Release flow clean:
  - merge to `main`
  - auto-tag next patch release
  - build matrix
  - publish GitHub release notes from PR metadata
- Migrations handled as lightweight workspace upgrades, not as heavy data-model drama.

### What `plan` should copy

- Repo-local file ownership.
- Minimal CLI surface.
- Skill installation model.
- Release automation pattern.
- Idempotent project upgrade model.

### What `plan` should not copy

- Memory and context ownership.
- Search/index/session systems.
- Anything that blurs planning into repo memory management.

## beads

### What works

- Serious multi-agent task coordination.
- Dependency-aware graph. `ready` queue especially strong.
- Collision-resistant IDs. Good for parallel work and branch merges.
- Convenience layers do not create separate systems. Example: `todo` is just a task shortcut.
- Epic progress and closure rules are operational, not vague.

### What `plan` should copy

- Optional dependency graph later.
- Ready-to-execute view later.
- "Convenience layer, not parallel system" philosophy.
- Safe IDs if and when planning objects need branch-safe creation.

### What `plan` should avoid

- Database-first architecture.
- Dolt dependency.
- Turning planning into issue tracking from day one.
- Huge operational surface area before core planning is excellent.

## get-shit-done

### What works

- Clear artifact pipeline. Project -> requirements -> roadmap -> phase context -> research -> plan -> execute -> verify.
- Strong gate philosophy. Plans are checked before execution.
- Plan tasks are explicit, executable, and verification-aware.
- Advanced workflows exist without removing happy-path simplicity.
- File-based state under `.planning/` is inspectable and local.

### What `plan` should copy

- Planning should generate execution-ready outputs, not fluffy notes.
- Verification expectations belong in planning artifacts.
- Advanced mode should be opt-in.
- Roadmap layer above detailed work is useful for big projects.

### What `plan` should avoid

- Owning context engineering as a product mission.
- Dozens of commands on day one.
- Phase-heavy workflow as the default mental model.
- Enterprise feeling or ceremony theater.

## Synthesis

Best direction for `plan`:

- Keep `brain` core flow as default mental model.
- Add a small amount of `gsd` discipline:
  - roadmap above execution units
  - plan checker mindset
  - explicit verification fields
- Borrow `beads` ideas only as optional power features:
  - dependencies
  - ready queues
  - branch-safe IDs

## Recommended product stance

`plan` should be:

- local-first
- markdown-first
- planning-only
- execution-aware but not execution-owning
- simple at the surface
- deep when needed

`plan` should not be:

- a memory system
- a context manager
- a hosted PM clone
- an issue tracker clone
- an enterprise process simulator

## Key conclusion

The best version of `plan` is not "brain without search" and not "mini-GSD" and not "beads for markdown users".

The best version is:

`brain`'s clean core flow
+ `gsd`'s rigor around plan quality and verification
+ later `beads`-style dependency power

That combination fits indie developers and still has room to grow into small-team workflows.
