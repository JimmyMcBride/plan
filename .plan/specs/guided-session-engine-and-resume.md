---
created_at: "2026-04-20T00:07:47Z"
epic: guided-session-engine-and-resume
project: plan
slug: guided-session-engine-and-resume
status: approved
target_version: v8
title: Guided Session Engine and Resume Spec
type: spec
updated_at: "2026-04-20T03:04:50Z"
---

# Guided Session Engine and Resume Spec

Created: 2026-04-20T00:07:47Z

## Why

Guided planning needs durable session state, not just better prompts.

## Problem

The current `plan` experience is pass-by-pass. Users can refine, challenge,
shape, or slice, but there is no durable guided session that remembers the
active stage, unfinished question cluster, or resume summary for a planning
chain.

## Goals

- add one active guided session per planning chain
- persist active stage, stage status, and cluster progress
- resume with a short summary so far, then reopen the active stage
- support `continue / refine / stop for now` as first-class session actions
- support switching between multiple guided chains in one repo

## Non-Goals

- hosted or shared session state
- cross-repo sessions
- stage-specific question sets
- repo scanning for planning documents

## Constraints

- all durable state stays in `.plan/.meta/`
- session state must work with existing brainstorm/epic/spec/story artifacts
- resume should favor clarity over hidden automation
- one chain should not overwrite another chain's session state

## Solution Shape

- add a guided-session record keyed to a planning chain
- track current stage, current cluster, stage statuses, and linked artifacts
- store a short summary snapshot and next-best-action text for resume
- store a repo-level pointer to the most recently active session
- use numbered CLI menus for `continue / refine / stop for now`

## Flows

1. User starts a guided brainstorm.
2. `plan` creates a guided session for that planning chain.
3. User answers one or more question clusters.
4. `plan` persists cluster progress and summary after each cluster.
5. User chooses `stop for now`.
6. Later, user resumes. `plan` shows a short summary so far, then reopens the
   active stage and continues from the next cluster.

## Data / Interfaces

- guided-session metadata in `.plan/.meta/`
- planning-chain identifier tied to brainstorm/epic/spec progression
- repo-level last-active pointer
- numbered session-action menu in CLI output

## Risks / Open Questions

- how much session state should be visible in normal CLI output versus only on resume
- how to keep session metadata small and durable as guided scope expands

## Rollout

- land session metadata and resume behavior first
- wire brainstorm flow to the session engine next
- add multi-session switching before deeper stage orchestration

## Verification

- starting a guided brainstorm creates a session record
- stopping mid-stage preserves the active stage and cluster progress
- resume shows a summary and continues the right stage
- multiple feature chains can coexist without clobbering each other

## Story Breakdown

## Analysis

### Missing Constraints

### Success Criteria Gaps

### Hidden Dependencies

### Risk Gaps

### What/Why vs How Leakage

### Recommended Revisions

## Checklist

## Resources

- [Epic](../epics/guided-session-engine-and-resume.md)
- [Brainstorm](../brainstorms/guided-planning-system.md)
- [Product Direction](../PRODUCT.md)

## Notes

Recommendation locked during brainstorming: one active session per planning
chain, plus a repo-level last-active pointer for quick resume.
