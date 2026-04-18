---
created_at: "2026-04-17T20:19:38Z"
project: plan
slug: brainstorm-challenge
spec: brainstorm-challenge
title: Brainstorm Challenge
type: epic
updated_at: "2026-04-17T20:19:38Z"
---

# Brainstorm Challenge

Created: 2026-04-17T20:19:38Z

## Outcome

Add a durable brainstorm challenge pass that pressure-tests ideas before they
harden into epics and specs.

## Why Now

`plan brainstorm refine` reduces ambiguity, but it does not yet push on risks,
overengineering, rabbit holes, or explicit no-gos. That leaves too much weak
shaping work to happen informally.

## Scope Boundary

- challenge sections in brainstorm notes
- an interactive `plan brainstorm challenge` pass
- resumable updates and idempotent note writes
- tests for the new challenge loop

Not in scope:

- automatic promotion to epics
- AI-provider-specific network calls
- new artifact types outside brainstorm notes

## Spec

- [Draft Spec](../specs/brainstorm-challenge.md)

## Resources

- [Product Direction](../PRODUCT.md)
- [Roadmap](../ROADMAP.md)

## Progress

- Target version: `v5`
- Status: planned

## Notes

The goal is to make shaping more adversarial in a useful way before teams invest
in specs and execution.
