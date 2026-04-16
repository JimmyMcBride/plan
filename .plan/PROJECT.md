# Project: plan

Created: 2026-04-16T05:33:06Z

## Vision

Build the best local-first planning tool for AI-assisted software projects.

`plan` should help indie developers and small teams turn rough ideas into
execution-ready specs and stories without PM theater, cloud lock-in, or context
management bloat.

## Principles

- Local-first and markdown-first.
- Planning only. No memory or context ownership.
- Specs are the contract.
- Simple default flow, deeper power later.
- Versions 1 through 3 stay focused on perfecting local planning before external integrations.

## Constraints

- All durable planning material lives in `.plan/`.
- No hosted dependency required for core workflows.
- No issue-tracker clone behavior in v1-v3.
- Integrations with GitHub/Jira/Linear are explicitly post-v3.

## Planning Rules

- Specs are the canonical execution contract.
- Stories are created only after spec approval.
- Stories should be execution-ready and verification-aware.

## Notes

- v1 builds the trustworthy local core.
- v2 adds planning rigor and adoption workflows.
- v3 adds local power features for bigger projects and small teams.
