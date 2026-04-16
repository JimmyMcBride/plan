---
created_at: "2026-04-16T05:46:34Z"
epic: skill-installer-and-release-delivery
project: plan
slug: harden-install-and-checksum-flow
spec: skill-installer-and-release-delivery
status: done
title: Harden install and checksum flow
type: story
updated_at: "2026-04-16T06:21:52Z"
---

# Harden install and checksum flow

Created: 2026-04-16T05:46:34Z

## Description


Tighten the install path so release downloads and checksum verification are trustworthy and predictable.
## Acceptance Criteria


- [ ] Install script resolves the latest release correctly

- [ ] Checksums are verified before install

- [ ] Fallback source-build behavior is clearly bounded
## Verification


- go test ./... && go build ./...
## Resources


- [Canonical Spec](../specs/skill-installer-and-release-delivery.md)
## Notes
