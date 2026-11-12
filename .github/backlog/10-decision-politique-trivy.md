---
title: "decision: policy for Trivy exceptions and unfixed vulnerabilities"
labels: ["type:decision", "pkg:image", "prio:normale", "taille:trait", "etat:a-valider"]
milestone: "v0.5.0 Container hardening"
assignee: PaulPerigault
---

## Context

`task docker:scan` fails on any HIGH or CRITICAL finding, with the vulnerability database as it is on the day. A finding with no upstream fix would block every pull request, and an ignore file added in a hurry would stay forever.

## Goal

A written rule says when `ignore-unfixed` or a `.trivyignore` entry is allowed, who adds it and when it expires.

## Acceptance criteria

- [ ] An ADR states the rule: at most one entry per CVE, with a reason, a link and an expiry date.
- [ ] `task docker:scan` and the CI `image` job read the same ignore file, or both use none.
- [ ] A check fails when an entry is past its expiry date (Go program in `dev/`, with a test).

## Package

image

## Size

trait

## Priority

normale

## Milestone

v0.5.0 Container hardening

## Out of scope

Choosing another scanner.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
