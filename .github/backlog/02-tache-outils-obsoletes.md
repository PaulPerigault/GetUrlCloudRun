---
title: "chore(tools): task tools:outdated lists pinned tools and images behind their latest release"
labels: ["type:tache", "pkg:ci", "prio:normale", "taille:page"]
milestone: "v0.2.0 Go upgrade and tooling baseline"
assignee: PaulPerigault
---

## Context

Everything is pinned on purpose: Go toolchain, the six tools of `tools/go.mod`, OpenTofu, the Google provider, Trivy and the two base images. Pinning only helps if somebody notices when a pin is old, and Dependabot does not cover OpenTofu, Trivy or the `.opentofu-version` file.

## Goal

One command prints each pin, its current version and the latest published one, read only, so the owner can decide an upgrade in one place.

## Acceptance criteria

- [ ] `task tools:outdated` prints a table: pin, current, latest, source, for Go, `tools/go.mod` tools, OpenTofu, provider, Trivy.
- [ ] The program is Go in `dev/`, standard library only, with a table test against a local `httptest` server (no public network in tests).
- [ ] `task check:docs` passes with the new task documented in `AGENTS.md`.

## Package

ci

## Size

page

## Priority

normale

## Milestone

v0.2.0 Go upgrade and tooling baseline

## Out of scope

Opening pull requests or editing a pin automatically. Dependabot stays the update mechanism where it exists.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
