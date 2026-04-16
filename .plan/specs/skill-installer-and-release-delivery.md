---
created_at: "2026-04-16T05:33:06Z"
epic: skill-installer-and-release-delivery
project: plan
slug: skill-installer-and-release-delivery
status: implementing
target_version: v1
title: Skill, Installer, and Release Delivery Spec
type: spec
updated_at: "2026-04-16T06:20:53Z"
---

# Skill, Installer, and Release Delivery Spec

Created: 2026-04-16T05:33:06Z

## Why

`plan` needs a polished install and release path from the beginning so the product can be used and iterated on outside this repo.

## Problem

Without a real installer, skill install path, and release automation, every user becomes a maintainer. That slows adoption and makes the workflow harder to trust.

## Goals

- provide a one-command install path
- install the `plan` skill globally or locally
- follow the `brain` release pattern from PR merge to release tag
- document maintainer expectations for release notes and verification

## Non-Goals

- supporting every package manager in `v1`
- building marketplace-specific integrations
- adding hosted update infrastructure

## Constraints

- installers must verify release checksums
- skill install must support the main local/global agent roots
- release notes should be generated from merged PR metadata

## Solution Shape

- ship `scripts/install.sh`
- ship embedded `skills/plan/`
- ship `plan skills install` and `plan skills targets`
- ship CI plus automatic tagging and GitHub release publishing

## Flows

1. Maintainer merges to `main`.
2. GitHub workflow tags the next patch release.
3. Build matrix publishes release assets and checksums.
4. User installs `plan`.
5. User installs the `plan` skill into the desired agent root.

## Data / Interfaces

- release archives per platform/arch
- checksums file per release
- skill manifest stored in installed skill directories

## Risks / Open Questions

- whether a PowerShell installer is needed during `v1`
- how broad the initial agent support list should remain

## Rollout

- ship core install path in `v1`
- validate the path with local dogfooding before first public release
- add broader distribution only after the local release flow is reliable

## Verification

- install script downloads and installs a valid binary
- `plan skills install` copies the skill bundle and writes a manifest
- CI passes `go test ./...` and `go build ./...`
- release workflows produce tagged assets and checksums

## Story Breakdown

- [ ] [Harden install and checksum flow](../stories/harden-install-and-checksum-flow.md)
- [ ] [Harden skill installation behavior](../stories/harden-skill-installation-behavior.md)
- [ ] [Validate release automation and maintainer docs](../stories/validate-release-automation-and-maintainer-docs.md)

## Resources

- [Epic](../epics/skill-installer-and-release-delivery.md)
- [Product Direction](../PRODUCT.md)

## Notes
