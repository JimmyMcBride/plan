---
created_at: "2026-04-16T05:46:34Z"
epic: skill-installer-and-release-delivery
project: plan
slug: harden-skill-installation-behavior
spec: skill-installer-and-release-delivery
status: todo
title: Harden skill installation behavior
type: story
updated_at: "2026-04-16T05:46:34Z"
---

# Harden skill installation behavior

Created: 2026-04-16T05:46:34Z

## Description


Make skill install targets, copied bundles, and manifests reliable across supported agent roots.
## Acceptance Criteria


- [ ] Global and local skill target resolution is correct

- [ ] Installed skill bundles include manifests and repair stale installs cleanly

- [ ] Skill install behavior is covered by tests
## Verification


- go test ./internal/skills
## Resources


- [Canonical Spec](../specs/skill-installer-and-release-delivery.md)
## Notes
