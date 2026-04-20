---
created_at: "2026-04-19T04:24:36Z"
project: plan
slug: guided-planning-system
status: active
title: Guided planning system
type: brainstorm
updated_at: "2026-04-20T00:08:43Z"
---

# Brainstorm: Guided planning system

Started: 2026-04-19T04:24:36Z

## Focus Question

How should plan become a genuinely guided, conversational co-planning system from brainstorming through story creation?
## Desired Outcome

Create a guided planning experience where plan and the user shape ideas together from rough vision to execution-ready stories.
## Constraints

- Keep the process streamlined rather than ceremonial.
- Do not just auto-fill goals, non-goals, constraints, and no-gos without real user back-and-forth.
- The system should feel like co-planning, not form filling.
## Open Questions

- No blocking open questions. Remaining implementation detail moves into epics and specs.
## Ideas

- Make brainstorming collaborative instead of a one-shot note capture.
- Ask clarifying questions to shape the user's mental vision before writing structured planning fields.
- Ask questions in small grouped clusters of 2-4 instead of one giant intake form.
- Keep the flow stage-by-stage, with an explicit offer to continue at the end of each stage.
- Let the agent challenge vague, bloated, or contradictory ideas and suggest cleaner paths forward.
- Move good-but-too-early ideas into roadmap parking so they are not lost.
- Use a default guided stage sequence of `vision -> clarify -> challenge -> epic -> spec -> stories`.
- When `plan` sees a gap, have it explain the issue, offer 2-3 possible ways forward, and ask the user how to proceed.
- End each stage with a short structured recap plus a choice to continue, refine, or stop for now.
- Replace and improve the current shaping commands instead of adding a separate guided command family.
- Make the gap-handling options short and opinionated: one recommended path plus 1-2 alternatives.
- Persist the active guided session automatically so users can leave and resume later.
- Park good-but-early ideas in `ROADMAP.md` so `plan` can surface them later as viable next bets.
- Start guided planning with user vision plus any relevant docs, links, and research context before shaping begins.
- Keep guided conversation as the default and only mode for now; no fast non-interactive path yet.
- On resume, show a short summary so far, then reopen the active stage and continue from there.
- Ask the user for relevant docs, links, and research context directly; do not auto-scan the repo for them.
- When upstream stages change, mark downstream stages as `needs review` instead of silently trusting them.
- Use one active guided session per planning chain, plus a repo-level `last active` pointer for easy resume.
- After a 2-4 question cluster, reflect back once against the whole cluster rather than after each answer.
- Use a numbered `continue / refine / stop for now` menu in the CLI for clarity and speed.
- When the user stops for now, save state and print a next-best-action summary for later.
## Raw Notes

The user wants plan to hold their hand from brainstorming to story creation. The system should ask what vision they have in their head, ask clarifying questions to shape that vision, and only then help structure the plan. Refinement should feel like a collaborative dialogue, not repeated prompting against prewritten sections.
Default should be user-input-first and back-and-forth. The agent can fill gaps when the user is okay with it, but that should be optional rather than the default planning style.
## Refinement

### Problem

`plan` is still too artifact-first. It can create and refine notes, but the
default experience still feels like filling planning sections instead of
co-planning with the user from rough vision through execution-ready stories.
### User / Value

Users get a guided planning partner that helps them discover the right shape of
their idea through conversation, challenge, and reflection before turning that
thinking into epics, specs, and stories.
### Appetite

Make the default planning experience more guided and conversational end-to-end,
but keep it streamlined and stage-based rather than turning it into a giant
wizard.
### Remaining Open Questions

- No blocking open questions. Remaining implementation detail moves into epics and specs.
### Candidate Approaches

- Replace one-shot brainstorm/refine flows with a guided stage session that asks
  small clusters of questions and reflects understanding back after each round.
- End each stage with a clear summary plus an explicit continue prompt into the
  next stage.
- Add structured challenge behavior that flags vagueness, bloat, contradiction,
  and mis-timed ideas, with suggested simplifications or roadmap parking.
- Replace and upgrade the current shaping commands so the default path becomes
  guided without splitting the product into two competing planning systems.
- Start each guided session by gathering the user's vision plus relevant docs,
  links, and research context before asking shaping questions.
- Ask the user directly for relevant docs, links, and research context rather
  than auto-scanning the repo.
- Keep user-input-first as the default, and when `plan` sees a gap, show the
  problem plus one recommended way forward and 1-2 alternatives, then ask the
  user how to proceed.
- Keep AI drafting as an explicit "want me to draft this for you?" escape hatch
  when the user wants speed.
- Persist active session/stage state automatically so users can stop and resume
  without losing the collaborative thread.
- On resume, show a short summary so far, reopen the active stage, and continue
  from the next question cluster.
- Use one active session per planning chain, with a repo-level last-active
  pointer for fast resume.
- After each question cluster, reflect back once at cluster level instead of
  interrupting after each answer.
- Use numbered CLI choices for `continue / refine / stop for now`.
- When the user stops for now, save state and print a short next-best-action
  summary for later.
- Allow backward jumps, but mark downstream stages as `needs review` whenever
  upstream changes reduce confidence in later work.
- Park good-but-early ideas in `ROADMAP.md` with value, reason parked, unlock
  condition, and source reference.
### Decision Snapshot

The right direction is a stage-by-stage co-planning system: question clusters,
reflection, pushback, and explicit handoff to the next stage, with user input
preferred by default and AI drafting as an opt-in assist. Each stage should end
with a short recap and a `continue / refine / stop for now` choice. The guided
system should replace and improve the current shaping commands rather than add
a parallel command family, and good-but-early ideas should go to roadmap
parking instead of being lost. Guided sessions should begin by gathering user
vision plus relevant supporting material, then resume with a short summary and
the active stage reopened automatically. Sessions should be chain-scoped, use
numbered stage menus, reflect once per question cluster, and ask the user for
supporting docs directly instead of auto-scanning the repo.
## Challenge

### Rabbit Holes

### No-Gos

- Do not default to auto-filling planning fields without first collaborating
  with the user.
- Do not turn the guided flow into a heavy wizard with too many steps at once.
- Do not force the user to accept AI-generated framing when they want to think
  through the idea themselves.
### Assumptions

- Users want guidance and pushback, not just faster template generation.
- Small question clusters will feel collaborative without becoming tedious.
- Explicit stage boundaries will keep the system streamlined while still
  feeling end-to-end.
- A short recap at the end of each stage will improve trust and reduce drift.
- Automatic session persistence will make longer planning loops practical.
- Most users will prefer guided conversation first, with no separate fast mode
  competing against it.
- Asking the user directly for supporting docs will keep context gathering
  intentional and avoid noisy auto-discovery.
### Likely Overengineering

- Adding too much hidden state or workflow machinery too early.
- Trying to automate every planning decision instead of keeping the user in the
  loop.
- Turning roadmap parking, challenge, and drafting into separate complex
  subsystems before the base guided loop is solid.
### Simpler Alternative

Keep the existing artifact model, but introduce a single guided session pattern
for each stage: ask a small question cluster, reflect back the current
understanding, challenge weak spots, then ask whether to continue.
