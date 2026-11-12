---
title: "ci(release): sign the image, attach an SBOM and a provenance attestation"
labels: ["type:tache", "pkg:image", "prio:haute", "taille:page"]
milestone: "v0.5.0 Container hardening"
assignee: PaulPerigault
---

## Context

The release workflow builds and attests the binaries only. The container image is built in CI for scanning and is never published, so a deployer cannot verify what they run.

## Goal

A tagged release publishes the image to a registry with a keyless signature, an SBOM and a build provenance, all verifiable by digest.

## Acceptance criteria

- [ ] `release.yml` builds the image once, pushes it by digest and signs it keylessly (OIDC), with actions pinned by full SHA.
- [ ] An SBOM (SPDX or CycloneDX) is generated from the pushed image and attached.
- [ ] The documented verification command (`cosign verify ...`) succeeds on a release candidate.
- [ ] `zizmor` and the `ci` job stay green; the release job is the only one with `id-token: write` and `packages: write`.

## Package

image

## Size

page

## Priority

haute

## Milestone

v0.5.0 Container hardening

## Out of scope

Deploying the image. Registries other than the one chosen in the decision step of this issue.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
