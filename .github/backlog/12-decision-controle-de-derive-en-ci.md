---
title: "decision: read-only drift check from CI through Workload Identity Federation"
labels: ["type:decision", "pkg:ci", "prio:basse", "taille:page", "etat:a-valider"]
milestone: "v0.7.0 CI/CD without apply"
assignee: PaulPerigault
---

## Context

ADR 0004 keeps CI free of any cloud credential, so drift between `infra/` and the real project is noticed only when the owner plans. A scheduled `plan -detailed-exitcode` would detect it, but it needs an identity with read rights in CI, which is a different risk from `apply` and still a decision.

## Goal

An ADR accepts or rejects a read-only drift check, with the exact scope of the identity if accepted.

## Acceptance criteria

- [ ] The ADR compares: no CI credential (today), a federated read-only identity restricted to this repository and the `main` branch, a manual `task` only.
- [ ] If accepted: the role list of the identity, its binding conditions and the workflow are written, and the workflow has no `apply` command and no write role (a test greps for both).
- [ ] If rejected: the reason is recorded and `infra/README.md` says how often the owner plans by hand.

## Package

ci

## Size

page

## Priority

basse

## Milestone

v0.7.0 CI/CD without apply

## Out of scope

Auto-remediation. Any apply from a pipeline, which stays forbidden.

## Definition of done

- [ ] Tests added, `task verify` green.
- [ ] Documentation up to date (README, ADR or journal).
- [ ] PR linked with `Closes #n`.
