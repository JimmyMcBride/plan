# Roadmap: plan

Created: 2026-04-16T05:33:06Z

## Overview

`plan` ships in three local-first versions. v1 proves the core workflow and
workspace model. v2 makes plans sharper, safer, and easier to maintain. v3 adds
power-user features that help larger projects and small teams stay local while
planning more complex work.

## v1: Local-First Core

Goal: Ship the trustworthy foundation for `plan` as a daily driver.

- [ ] Core Workspace and Artifact System
- [ ] Spec-Driven Planning Workflow
- [ ] Skill, Installer, and Release Delivery

Summary:
- establish `.plan/` as the durable workspace
- make `brainstorm -> epic -> spec -> story` work cleanly
- ship install, skill, docs, and release flow so `plan` is usable outside this repo

## v2: Planning Rigor

Goal: Make plans more reliable and easier to scale without bloating the model.

- [ ] Roadmap and Portfolio Planning
- [ ] Plan Quality and Verification Engine
- [ ] Workspace Adoption, Update, and Migration

Summary:
- make roadmap planning first-class
- tighten spec and story quality with explicit checks
- support repo adoption and safe workspace evolution

## v3: Local Power

Goal: Add the power features that make `plan` strong on big projects while staying local-first.

- [ ] Dependency Graph and Ready Work
- [ ] Brain Interop and Planning Imports
- [ ] Power-User Local Workflows

Summary:
- model dependencies and surface ready work
- interoperate cleanly with `brain`
- support richer local workflows for advanced users and small teams

## Ordering Notes

- v1 proves the core file model and daily workflow.
- v2 improves rigor before widening scope.
- v3 adds power features after the core is stable.
- External integrations stay out of scope for the first three versions.

## Parking Lot

- GitHub sync/export workflows
- Jira or Linear adapters
- hosted dashboards
- database-backed multi-user mode
- any memory, retrieval, or context-engineering features
